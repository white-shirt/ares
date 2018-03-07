package interceptor

import (
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/sevenNt/ares/metric"
	"github.com/sevenNt/ares/server/yell"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Metrics wraps sets of metric.
var Metrics map[string]*Metric

// DefaultMetric is default metric interceptor.
var DefaultMetric = &Metric{
	interval: time.Second * 10,
	nameFn: func(method, prefix string) string {
		name := strings.Replace(method, "/", "_", -1)
		name = strings.Replace(name, ".", "_", -1)
		name = prefix + "_" + name
		name = strings.Replace(name, ":", "", -1)
		name = strings.Replace(name, "@", "", -1)
		name = strings.Replace(name, "#", "", -1)
		return name
	},
}

// Metric wraps metric interceptor.
type Metric struct {
	Base
	pusher   *metric.PrometheusPusher
	name     string                             // 构建metric name, 默认 method+path
	nameFn   func(method, prefix string) string //name处理方法
	prefix   string
	interval time.Duration //心跳消息时间

	hs map[string]metrics.Histogram
}

// NewMetric constructs a new metric interceptor.
func NewMetric(interval time.Duration, prefix string, p *metric.PrometheusPusher) *Metric {
	return &Metric{
		prefix:   "_" + prefix,
		nameFn:   DefaultMetric.nameFn,
		interval: interval,
		pusher:   p,
		hs:       make(map[string]metrics.Histogram),
	}
}

// SetNameFn sets name function of metric.
func (m *Metric) SetNameFn(f func(string, string) string) {
	m.nameFn = f
}

// watch 监听心跳消息，数据上报。
func (m *Metric) watch() {
	ticker := time.Tick(m.interval)
	for {
		select {
		case <-ticker:
			for n, h := range m.hs {
				m.pusher.AddCollector(n+"_tps", m.prefix+"_tps",
					"The request numbers of a specific URL",
					float64(h.Count()))

				m.pusher.AddCollector(n+"_p99", m.prefix+"_p99",
					"The request numbers of a specific URL",
					h.Percentile(0.99))
				h.Clear()
			}
		}
	}
}

// LoadServiceInfo loads grpc server.
func (m *Metric) HookBeforeServe(s *yell.Server) {
	for fm, info := range s.GetServiceInfo() {
		for _, method := range info.Methods {
			name := m.nameFn(fm+"_"+strings.Title(method.Name), m.prefix)
			m.hs[name] = metrics.NewRegisteredHistogram(m.name, nil, metrics.NewExpDecaySample(1028, 0.015)) //初始化接口性能数据
		}
	}
}

// UnaryIntercept implements UnaryIntercept function of Interceptor.
func (m *Metric) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	go m.watch() //并行监听心跳消息，数据上报
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		beg := time.Now()
		resp, err = handler(ctx, req)
		name := m.nameFn(strings.Replace(info.FullMethod, "/", "", 1), m.prefix)
		if h, ok := m.hs[name]; ok {
			h.Update(int64(time.Since(beg)))
		}
		return
	}
}

// StreamIntercept implements StreamIntercept function of Interceptor.
func (m *Metric) StreamServerIntercept() grpc.StreamServerInterceptor {
	go m.watch() //并行监听心跳消息，数据上报
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		beg := time.Now()
		err = handler(srv, stream)
		name := m.nameFn(strings.Replace(info.FullMethod, "/", "", 1), m.prefix)
		if h, ok := m.hs[name]; ok {
			h.Update(int64(time.Since(beg)))
		}
		return
	}
}
