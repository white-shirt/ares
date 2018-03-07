package interceptor

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sevenNt/ares/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	// LocalIPKey port, sid key in context
	LocalIPKey = "local_ip"

	// LocalPortKey ...
	LocalPortKey = "local_port"

	// LocalSID ...
	LocalSID = "local_sid"
	// HeaderTraceStatus for trace_status
	HeaderTraceStatus = "tstt"
	// HeaderTid refers to unique request id
	HeaderTid = "t"
	// HeaderPid refers to unique service id
	HeaderPid = "s"
	// HeaderPort refers to last level port
	HeaderPort = "p"
	// RemoteIP refers to last level IP
	RemoteIP = "i"
	// HeaderAid refers to application id
	HeaderAid = "a"
	// HeaderAuth refers to auth
	HeaderAuth = "au"
)

const (
	defaultTraceStatus = "0"

	errTraceStatus = "1"

	forceTraceStatus = "2"
)

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

// TraceWrite Nsq消息相关
type TraceWrite struct {
	Base
	msgCh     chan []byte   //发送nsq的链路消息
	stopCh    chan struct{} //停止发送
	localAddr string
	logPath   string
	forceLog  bool
}

// Trace 每次请求的数据
type Trace struct {
	Base
	startTime   time.Time //请求的起始时间
	name        string    //当前请求的接口名称
	md          metadata.MD
	traceStatus string
}

var twrite *TraceWrite

// NewTraceWrite ..., localAddr 项目地址，eg: 127.0.0.0:50001
func NewTraceWrite(localAddr string, logPath string) *TraceWrite {
	twrite = &TraceWrite{
		msgCh:     make(chan []byte, 1000),
		localAddr: localAddr,
		logPath:   logPath,
	}
	go twrite.writeTrace()
	return twrite
}

// UnaryIntercept implements UnaryIntercept function of Interceptor.
func (t *TraceWrite) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		trace := Trace{traceStatus: "0"}
		t.forceLog = false
		md, _ := metadata.FromIncomingContext(ctx)
		var localAddr = strings.Split(twrite.localAddr, ":")
		if len(localAddr) != 2 {
			fmt.Println("local addr format may be wrong", localAddr)
			localAddr = []string{"0.0.0.0", "0"}
		}

		var peerAddr []string
		if p, ok := peer.FromContext(ctx); ok {
			peerAddr = strings.Split(p.Addr.String(), ":")
			if len(peerAddr) != 2 {
				fmt.Println("remote addr format may be wrong", localAddr)
				peerAddr = []string{"0.0.0.0", "0"}
			}
		} else {
			fmt.Println("get remote addr wrong")
			peerAddr = []string{"0.0.0.0", "0"}
		}
		md = metadata.Join(md, metadata.Pairs(LocalIPKey, localAddr[0], LocalPortKey, localAddr[1], RemoteIP, peerAddr[0]))
		trace.startTrace(info, md)
		ctx = metadata.NewOutgoingContext(ctx, trace.setNetContext())
		trace.traceStatus = mdValueGet(md, HeaderTraceStatus)
		if trace.traceStatus == "" || trace.traceStatus == defaultTraceStatus {
			if t.forceLog {
				trace.traceStatus = forceTraceStatus
			} else {
				trace.traceStatus = defaultTraceStatus
			}
		}
		res, err = handler(ctx, req)
		trace.endTrace()
		return
	}
}

// SetForceLogTrue force logging.
func (t *TraceWrite) SetForceLogTrue() {
	t.forceLog = true
}

// Stop sends msg to stop channel.
func (t *TraceWrite) Stop() {
	t.stopCh <- struct{}{}
}

// 写trace日志 todo 待完善
func (t *TraceWrite) writeTrace() {
	for c := range t.msgCh {
		t.writeLog(string(c) + "\n")
	}
}

func (t *TraceWrite) writeLog(msg string) {
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
	m, _ := file.Seek(0, os.SEEK_END)
	_, err = file.WriteAt([]byte(msg), m)

	if err != nil {
		fmt.Println("write error", err.Error())
	}
}

func (t *Trace) startTrace(info *grpc.UnaryServerInfo, md metadata.MD) {
	t.name = strings.Replace(info.FullMethod, "/", ".", -1)
	t.startTime = time.Now()
	t.md = md
	t.sid()
}

func (t *Trace) endTrace() {
	msg := t.createMsg()
	if msg != nil {
		select {
		case <-twrite.stopCh:
			close(twrite.msgCh)
		case twrite.msgCh <- msg:
		}
	}
}

func (t *Trace) createMsg() []byte {
	traceID := mdValueGet(t.md, HeaderTid)

	if traceID == "" {
		fmt.Println("traceID is empty")
		return nil
	}

	exception := ""

	if len(traceID) != 32 {
		fmt.Println("wrong traceID: ", traceID)
		return nil
	}
	sampleStr := traceID[31:] //tid后1个字节用于表示是否采样
	i, err := strconv.ParseInt(sampleStr, 16, 0)
	if err != nil {
		fmt.Println("fail to parse sample value:", err)
		return nil
	}
	if i != 1 {
		if t.traceStatus == defaultTraceStatus {
			fmt.Println(`[xzl]i----------->`, i)
			return nil
		}
	} else {
		if t.traceStatus == forceTraceStatus {
			t.traceStatus = defaultTraceStatus
		}
	}

	msg := msgStruct{
		Pid:         mdValueGet(t.md, HeaderPid),
		Tid:         traceID,
		Sid:         mdValueGet(t.md, LocalSID),
		Aid:         mdValueGet(t.md, HeaderAid),
		RemoteAddr:  remoteAddr(t.md),
		LocalAddr:   localAddr(t.md),
		Name:        t.name,
		StartTime:   strconv.FormatInt(t.startTime.UnixNano()/1000, 10),
		Value:       time.Since(t.startTime) / 1000,
		Exception:   exception,
		TraceStatus: t.traceStatus,
	}

	msgByte, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return msgByte
}

func (t *Trace) sid() (sid string) {
	sid = util.GenerateID()
	t.md = metadata.Join(t.md, map[string][]string{LocalSID: []string{sid}})
	return
}

func (t *Trace) setNetContext() metadata.MD {
	md := metadata.Pairs(
		HeaderTid, mdValueGet(t.md, HeaderTid),
		HeaderAid, mdValueGet(t.md, HeaderAid),
		HeaderPid, mdValueGet(t.md, LocalSID),
		HeaderPort, mdValueGet(t.md, LocalPortKey),
		HeaderAuth, md5Auth(mdValueGet(t.md, LocalSID), mdValueGet(t.md, HeaderTid)),
	)
	return md
}

var md5Auth = defaultMd5Auth

//计算md5 auth
func defaultMd5Auth(sid string, tid string) (md5Str string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(sid + "dytraceauth" + tid))
	cipherStr := md5Ctx.Sum(nil)
	md5Str = hex.EncodeToString(cipherStr)
	return
}

func mdValueGet(md metadata.MD, name string) (value string) {
	value = ""
	if len(md[name]) > 0 {
		value = md[name][0]
	}
	return
}

func remoteAddr(md metadata.MD) string {
	remotePort := mdValueGet(md, HeaderPort)
	remoteIP := mdValueGet(md, RemoteIP)

	remoteAddr, err := util.Addr2Hex(net.JoinHostPort(remoteIP, remotePort))
	if err != nil {
		fmt.Println("fail to transform localAddr to hex, error:", err)
		return ""
	}
	return remoteAddr
}

func localAddr(md metadata.MD) string {
	localIP := mdValueGet(md, LocalIPKey)
	localPort := mdValueGet(md, LocalPortKey)
	if localIP == "" || localPort == "" {
		return ""
	}

	localAddr, err := util.Addr2Hex(net.JoinHostPort(localIP, localPort))
	if err != nil {
		fmt.Println("fail to transform localAddr to hex, error:", err)
		return ""
	}
	return localAddr
}

func isFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
