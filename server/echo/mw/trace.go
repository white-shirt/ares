package mw

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sevenNt/ares/server/echo"
	"github.com/sevenNt/ares/util"
	"google.golang.org/grpc/metadata"
)

const (
	// HeaderTid refers to unique request id
	HeaderTid = "t"
	// HeaderSid refers to unique service id
	HeaderSid = "s"
	// HeaderPort refers to last level port
	HeaderPort = "p"
	// HeaderTraceStatus for trace_status
	HeaderTraceStatus = "tstt"
	// HeaderAid refers to application id
	HeaderAid = "a"
)

const (
	defaultTraceStatus = "0"

	errTraceStatus = "1"

	forceTraceStatus = "2"
)

// Trace defines the config for Trace middleware.
type Trace struct {
	Base
	Setter      func(http.Header)
	stopCh      chan struct{}
	msgCh       chan []byte
	startTime   time.Time
	traceStatus string
	name        string
	localAddr   string
	logPath     string
	forceLog    bool
	localSid    string
}

type msgStruct struct {
	Exception   string        `json:"exception"`
	LocalAddr   string        `json:"local_addr"`
	Pid         string        `json:"pid"`
	Sid         string        `json:"sid"`
	TraceStatus string        `json:"trace_status"`
	Name        string        `json:"name"`
	Value       time.Duration `json:"value"`
	Aid         string        `json:"aid"`
	Tid         string        `json:"tid"`
	RemoteAddr  string        `json:"remote_addr"`
	StartTime   string        `json:"start_time"`
}

// NewTrace constructs a new Trace instance.
func NewTrace(localAddr string, logPath string) *Trace {
	t := &Trace{
		stopCh:    make(chan struct{}),
		msgCh:     make(chan []byte, 1000),
		localAddr: localAddr,
		logPath:   logPath,
		forceLog:  false,
	}
	return t
}

// Func ...
func (t *Trace) Func() echo.MiddlewareFunc {
	go t.writeTrace()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			t.localSid = util.GenerateID()
			t.traceStatus = c.Request().Header().Get(HeaderTraceStatus)
			if t.traceStatus == "" || t.traceStatus == defaultTraceStatus {
				if t.forceLog {
					t.traceStatus = forceTraceStatus
					c.Response().Header().Set("traceStatus", forceTraceStatus)
				} else {
					t.traceStatus = defaultTraceStatus
				}
			}

			t.HookRoute("", c.Request().URL.Path())
			c.Context = metadata.NewOutgoingContext(c.Context, t.setNetContext(c))
			err := next(c)
			t.endTrace(c)
			return err
		}
	}
}

// SetForceLogTrue force logging.
func (t *Trace) SetForceLogTrue() {
	t.forceLog = true
}

// Stop ...
func (t *Trace) Stop() {
	t.stopCh <- struct{}{}
}

func (t *Trace) writeTrace() {
	for c := range t.msgCh {
		t.writeLog(string(c) + "\n")
	}
}

func (t *Trace) endTrace(c *echo.Context) {
	msg := t.createMsg(c)
	if msg != nil {
		select {
		case <-t.stopCh:
			close(t.msgCh)
		case t.msgCh <- msg:
		}
	}
}

func (t *Trace) setNetContext(c *echo.Context) metadata.MD {
	var localAddr = strings.Split(t.localAddr, ":")
	if len(localAddr) != 2 {
		fmt.Println("local addr format may be wrong", localAddr)
		localAddr = []string{"0.0.0.0", "0"}
	}
	md := metadata.Pairs(
		HeaderTid, c.Request().Header().Get(HeaderTid),
		HeaderAid, c.Request().Header().Get(HeaderAid),
		HeaderSid, t.localSid,
		HeaderPort, localAddr[1],
		HeaderTraceStatus, t.traceStatus,
	)
	return md
}

/*
{"pid":"81907883b9a025f8","tid":"975f03ed0ff494561fb274dd482bd9ce09","value":5133,"sid":"a0c640b86899b6ff","start_time":"1500433834965387",
"remote_addr":"0a0133160050","aid":"0","local_addr":"c0a847880050","name":"\/api\/thirdpart\/live",
"exception":"{\"code\":1501,\"msg\":\"limit\\u53c2\\u6570\\u975e\\u6cd5\"}","trace_status":"1"}
*/

func remoteAddr(c *echo.Context) string {
	remotePort := c.Request().Header().Get(HeaderPort)
	var remoteAddrPart = strings.Split(c.Request().RemoteAddr, ":")
	var remoteIP string
	if len(remoteAddrPart) > 0 {
		remoteIP = remoteAddrPart[0]
	}

	remoteAddr, err := util.Addr2Hex(net.JoinHostPort(remoteIP, remotePort))
	if err != nil {
		fmt.Println("fail to transform localAddr to hex, error:", err)
		return ""
	}
	return remoteAddr
}

func (t *Trace) createMsg(c *echo.Context) []byte {
	traceID := c.Request().Header().Get(HeaderTid)
	if traceID == "" {
		fmt.Println("[Trace]traceID is empty")
		return nil
	}

	if len(traceID) != 32 {
		fmt.Println("wrong traceID: ", traceID)
		return nil
	}
	traceStatus, exception := t.exceptionInfo(c)
	sampleStr := traceID[31:] //tid后1个字节用于表示是否采样
	i, err := strconv.ParseInt(sampleStr, 16, 0)

	if err != nil {
		fmt.Println("fail to parse sample value:", err)
		return nil
	}

	if i != 1 {
		if traceStatus == defaultTraceStatus {
			fmt.Println(`[xzl]i----------->`, i)
			return nil
		}
	} else {
		if traceStatus == forceTraceStatus {
			traceStatus = defaultTraceStatus
		}
	}

	hexAddr, err := util.Addr2Hex(t.localAddr)
	if err != nil {
		fmt.Println("fail to transform localAddr to hex, error:", err)
	}
	msg := msgStruct{
		Pid:         c.Request().Header().Get(HeaderSid),
		Tid:         traceID,
		Sid:         t.localSid,
		Aid:         c.Request().Header().Get(HeaderAid),
		RemoteAddr:  remoteAddr(c),
		LocalAddr:   hexAddr,
		Name:        t.name,
		StartTime:   strconv.FormatInt(t.startTime.UnixNano()/1000, 10),
		Value:       time.Since(t.startTime) / 1000,
		Exception:   exception,
		TraceStatus: traceStatus,
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return msgByte
}

// HookRoute hooks route with provided method and path.
func (t *Trace) HookRoute(method string, path string) {
	t.name = strings.Replace(path, "/", ".", -1)
	t.startTime = time.Now()
}

// Clone ...
func (t Trace) Clone() (echo.Middleware, bool) {
	tt := t
	return &tt, true
}

func (t *Trace) writeLog(msg string) {
	var file *os.File
	var err error
	filePath := t.logPath + "/hrpc" + time.Now().Format("2006-01-02") + ".log"
	if !isFileExist(filePath) {
		if !isFileExist(t.logPath) {
			if err = os.MkdirAll(t.logPath, 0755); err != nil {
				fmt.Println("MkdirAll error", err.Error())
				return
			}
		}

		file, err = os.Create(filePath)
		if err != nil {
			fmt.Println("Createfile error", err.Error())
			return
		}

	} else {
		file, err = os.OpenFile(filePath, os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("openfile error", err.Error())
			return
		}
	}
	defer file.Close()
	m, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		fmt.Println("write error1", err.Error())
		return
	}
	_, err = file.WriteAt([]byte(msg), m)

	if err != nil {
		fmt.Println("write error2", err.Error())
	}
}

func isFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (t *Trace) exceptionInfo(c *echo.Context) (string, string) {
	// Ret 为返回json数据结构体
	type Ret struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}

	retJSON := c.Get("TraceException")
	traceStatus := t.traceStatus
	exceptionInfo := ""
	if retJSON != nil { //如果有exception则必定写当前链路信息， 并在response header中设置status
		traceStatus = errTraceStatus
		v := retJSON.(Ret)
		if b, err := json.Marshal(v); err == nil {
			exceptionInfo = string(b)
		}
	}
	return traceStatus, exceptionInfo
}
