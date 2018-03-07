package ares

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/flag"
	"github.com/sevenNt/ares/registry"
	"github.com/sevenNt/ares/server"
	"github.com/sevenNt/ares/server/echo"
	"github.com/sevenNt/ares/server/yell"
	"github.com/sevenNt/ares/worker"
	"github.com/sevenNt/hera"
	"github.com/sevenNt/wzap"
)

var (
	app *App
)

// App allows users to configure the application.
type App struct {
	options Options
	*cron.Cron
	workers map[string]worker.Worker
	servers map[string]server.Server // map[label]server.Server
	//listeners   map[string]net.Listener
	wg          sync.WaitGroup
	sigChan     chan os.Signal
	errs        chan error
	sigHandlers map[os.Signal]func(os.Signal) error
	running     bool
	defers      []func()

	serverOpts map[string]server.Options // map[label]server.Server
}

// NewAPP constructs a new app from provided options.
func NewAPP(opts ...Option) *App {
	flag.Parse()
	app = &App{
		Cron:       cron.New(),
		sigChan:    make(chan os.Signal, 1),
		servers:    make(map[string]server.Server),
		workers:    make(map[string]worker.Worker),
		defers:     make([]func(), 0),
		serverOpts: make(map[string]server.Options), // scheme:host:port
	}

	app.loadOptions(opts...)
	app.initLogger()
	return app
}

func (app *App) loadOptions(opts ...Option) {
	// 1.0 从NewApp方法参数中载入配置
	var options = defaultOptions
	for _, option := range opts {
		option(&options)
	}

	// 2.0 从配置文件中载入配置
	if mode := hera.GetString("app.mode"); mode != "" {
		options.mode = mode
	}
	if debug := hera.GetBool("app.debug"); debug {
		options.debug = debug
	}
	if systemd := hera.GetBool("app.systemd"); systemd {
		options.systemd = systemd
	}
	app.options = options
	application.SetLabel(app.options.version, app.options.mode)
}

// SetMaxProcs sets maximum number of CPUs.
func (app *App) SetMaxProcs(num int) {
	runtime.GOMAXPROCS(num)
}

// Run runs application.
func (app *App) Run() {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, true)
			fmt.Printf("%s %s", err, stack[:length])
			app.unregister()
			time.Sleep(time.Millisecond)
		}
	}()

	app.initRuntime()
	app.hookSignals()
	app.initRegistry()
	app.initMetric()
	app.run()
}

// Defer adds defer functions which deferring during application shutting down.
func (app *App) Defer(fns ...func()) {
	app.defers = append(app.defers, fns...)
}

func (app *App) initRegistry() {
	if hera.Get("app.registry.etcd") != nil {
		registry.InitRegistry(registry.NewETCDRegistry(
			registry.WithAddrs(hera.GetStringSlice("app.registry.etcd.endpoints")),
			registry.WithTimeout(hera.GetDuration("app.registry.etcd.timeout")),
			registry.WithSecure(hera.GetBool("app.registry.etcd.secure")),
		))
	}
}

func (app *App) initRuntime() {
	if num := hera.GetInt("app.maxproc"); num != 0 {
		app.SetMaxProcs(num)
	}
}

func (app *App) initMode() {
	if mode := hera.GetString("app.mode"); mode != "" {
		app.options.mode = mode
	} else {
		panic("undefined app mode")
	}
}

func (app *App) initLogger() {
	logs := hera.GetStringMap("app.logger")
	for l := range logs {
		outputs := hera.GetStringMap("app.logger." + l)
		outputs["name"] = l
		wzap.Register(l, wzap.New(wzap.WithOutputKVs([]interface{}{outputs})))
	}
}

// AddWorker adds a new worker to application.
func (app *App) AddWorker(label string, w worker.Worker) {
	app.workers[label] = w
}

func (app *App) run() {
	app.running = true

	app.Cron.Start()
	defer app.Cron.Stop()

	// run workers
	for _, wrk := range app.workers {
		app.wg.Add(1)
		go func(wrk worker.Worker) {
			defer app.wg.Done()
			wrk.Run()
		}(wrk)
	}

	// run server
	for name, srv := range app.servers {
		if srv == nil || name == "" {
			wzap.Errorf("[service] invalid service or empty name")
			continue
		}

		srv.Mux()
		app.wg.Add(1)
		go func(srv server.Server, name string) {
			defer app.wg.Done()
			wzap.Infof("[service] srv %#v", srv)
			if err := srv.Serve(); err != nil {
				wzap.Errorf("[service] srv.Serve failed: %s", err)
			}
		}(srv, name)
	}

	// 等待启动，并检测超时. 同时也延迟注册,保证服务可用
	// NOTE(lvchao) 直接检测端口占用更可靠，但在graceful restart的时候
	// 端口会一直被占用，检测端口又会变得不可靠。暂时使用最简单的sleep方式
	// NOTE(lvchao) 需要注意的是 srv.Mux如果阻塞或者长时任务，sleep方式会恶化以上情况
	time.Sleep(time.Second * 3)
	for _, srv := range app.servers {
		if !srv.IsRunning() {
			wzap.Fatalf("service %s start timeout", srv.Addr())
			os.Exit(1)
		}
	}
	app.register()
	app.wg.Wait()
	app.running = false
	wzap.Info("app exit")
}

func (app *App) shutdown() {
	wzap.Warn("[ares] shutdown...")
	app.unregister()
	var wg sync.WaitGroup
	for name, srv := range app.servers {
		wg.Add(1)
		go func(name string, srv server.Server) {
			defer wg.Done()
			srv.GracefulStop()
		}(name, srv)
	}
	for _, wrk := range app.workers {
		wg.Add(1)
		go func(wrk worker.Worker) {
			defer wg.Done()
			wrk.GracefulStop()
		}(wrk)
	}
	wg.Wait()
	wzap.Warn("[ares] shutdown servers/workers done!")
	wzap.Warn("[cleanup]... begin.")
	for i := len(app.defers); i >= 1; i-- {
		deferFn := app.defers[i-1]
		deferFn()
	}

	// TODO stop application metric if it is not nil
	wzap.Warn("[cleanup]... done!")
	time.Sleep(time.Millisecond * 500)
	os.Exit(0)
}

func (app *App) terminate() {
	app.unregister()
	for _, srv := range app.servers {
		srv.Stop()
	}
	for _, wrk := range app.workers {
		wrk.Stop()
	}
	for i := len(app.defers); i >= 1; i-- {
		deferFn := app.defers[i-1]
		deferFn()
	}
	time.Sleep(time.Millisecond * 500)
	os.Exit(1)
}

// Serve serves provided servers.
func (app *App) Serve(srv server.Server, opts ...server.Option) {
	v := reflect.ValueOf(srv).Elem()
	t := reflect.TypeOf(srv).Elem()
	if t.Kind() != reflect.Struct {
		panic("reflect error: service struct")
	}

	var options server.Options
	for _, opt := range opts {
		opt(&options)
	}

	// appname:scheme:host:port
	//label := fmt.Sprintf("%s:%s:%s:%d", app.Label(), srv.Scheme(), options.host, options.port)
	label := fmt.Sprintf("%s:%s:%s", srv.Scheme(), app.options.Label(), options.Addr())

	listener, err := net.Listen("tcp4", options.Addr())
	if err != nil {
		panic(err)
	}

	if tf, ok := t.FieldByName("Server"); ok {
		switch tf.Type {
		case reflect.TypeOf(&echo.Server{}):
			v.FieldByName("Server").Set(reflect.ValueOf(echo.NewServer(listener, opts...)))
		case reflect.TypeOf(&yell.Server{}):
			v.FieldByName("Server").Set(reflect.ValueOf(yell.NewServer(listener, opts...)))
		}
	}

	if _, ok := app.servers[label]; ok {
		panic("service name " + label + " is duplicate")
	}

	app.servers[label] = v.Addr().Interface().(server.Server)
	app.serverOpts[label] = options
}

func checkAddr(addr string) {
	// check splitting addr
	host, port, e := net.SplitHostPort(addr)
	if e != nil {
		wzap.Fatalf("invalid addr:%s, error: %s", addr, e.Error())
	}

	// check host
	if ip := net.ParseIP(host); host != "" && ip == nil {
		wzap.Fatalf("invalid host:%s", host)
	}

	// check port
	_, e = strconv.ParseUint(port, 10, 16)
	if e != nil {
		wzap.Fatalf("invalid port:%q, error: %s", port, e.Error())
	}
}
