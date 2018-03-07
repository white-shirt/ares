package metric

import (
	"os"
	"runtime"
	"time"

	"github.com/sevenNt/wzap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/process"
)

type appMetricCollector struct {
	goroutinesDesc *prometheus.Desc
	threadsDesc    *prometheus.Desc
	heapDesc       *prometheus.Desc
	heapIdleDesc   *prometheus.Desc
	heapInuseDes   *prometheus.Desc
	cpu            *prometheus.Desc
	connNum        *prometheus.Desc
	bytesIn        *prometheus.Desc
	bytesOut       *prometheus.Desc
	timeStamp      *prometheus.Desc
	fds            *prometheus.Desc
	p              *process.Process
}

//m.pusher.AddCollector(m.name+"_hbs", m.prefix+"_hbs",
//"The request numbers of a specific URL",
//float64(time.Now().UnixNano()))

// NewAppMetricCollector ...
func NewAppMetricCollector(appLabels map[string]string) prometheus.Collector {
	c := &appMetricCollector{
		goroutinesDesc: prometheus.NewDesc(
			"app_goroutines",
			"Number of goroutines that currently exist.",
			nil, nil),
		threadsDesc: prometheus.NewDesc(
			"app_thread",
			"Number of OS threads created.",
			nil, nil),
		heapDesc: prometheus.NewDesc(
			"app_memstats_heap_alloc_bytes",
			"Number of heap bytes allocated and still in use.",
			nil, nil,
		),
		heapIdleDesc: prometheus.NewDesc(
			"app_memstats_heap_idle_bytes",
			"Number of heap bytes allocated and still in use.",
			nil, nil,
		),
		heapInuseDes: prometheus.NewDesc(
			"app_memstats_heap_inuse_bytes",
			"Number of heap bytes waiting to be used..",
			nil, nil,
		),
		cpu: prometheus.NewDesc(
			"app_cpu",
			"CPU percent of current pid",
			nil, nil,
		),
		connNum: prometheus.NewDesc(
			"app_connection_number",
			"Number of connections(TCP,UDP,UNIX)",
			nil, nil,
		),
		bytesIn: prometheus.NewDesc(
			"app_bytes_in",
			"Number of bytes sent",
			nil, nil,
		),
		bytesOut: prometheus.NewDesc(
			"app_bytes_out",
			"Number of bytes received",
			nil, nil,
		),
		timeStamp: prometheus.NewDesc(
			"app_heart_beat_time",
			"last time receive from app",
			nil, appLabels,
		),
		fds: prometheus.NewDesc(
			"app_fds",
			"Number of file descriptors",
			nil, nil,
		),
	}

	pid := os.Getpid()
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		wzap.Info("[metric]", "error", err)
		return nil
	}
	if _, err := p.Percent(0); err != nil {
		wzap.Info("[metric]", "error", err)
	}
	c.p = p
	return c
}

// SetMetric is to encapsulate the prometheus MustNewMetric
func SetMetric(ch chan<- prometheus.Metric,
	desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) {
	m, err := prometheus.NewConstMetric(desc, valueType, value, labelValues...)
	if err == nil {
		ch <- m
	} else {
		wzap.Info("[metric]", "error ", err)
	}
}

// Describe returns all descriptions of the collector.
func (c *appMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.goroutinesDesc
	ch <- c.threadsDesc
	ch <- c.heapDesc
	ch <- c.heapIdleDesc
	ch <- c.heapInuseDes
	ch <- c.cpu
	ch <- c.connNum
	ch <- c.bytesIn
	ch <- c.bytesOut
	ch <- c.fds
	ch <- c.timeStamp
}

// Collect returns the current state of all metrics of the collector.
func (c *appMetricCollector) Collect(ch chan<- prometheus.Metric) {
	SetMetric(ch, c.goroutinesDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	n, _ := runtime.ThreadCreateProfile(nil)
	SetMetric(ch, c.threadsDesc, prometheus.GaugeValue, float64(n))

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	SetMetric(ch, c.heapDesc, prometheus.GaugeValue, float64(ms.HeapAlloc))
	SetMetric(ch, c.heapIdleDesc, prometheus.GaugeValue, float64(ms.HeapIdle))
	SetMetric(ch, c.heapInuseDes, prometheus.GaugeValue, float64(ms.HeapInuse))

	cpu, err := c.p.Percent(0)
	if err != nil {
		wzap.Info("[metric]", "error", err)
	}
	SetMetric(ch, c.cpu, prometheus.GaugeValue, cpu)

	conRes, err := c.p.Connections()
	if err != nil {
		wzap.Info("[metric]", "error", err)
	}
	SetMetric(ch, c.connNum, prometheus.GaugeValue, float64(len(conRes)))

	if netStat, err := c.p.NetIOCounters(false); err == nil && len(netStat) == 1 {
		SetMetric(ch, c.bytesIn, prometheus.GaugeValue, float64(netStat[0].BytesRecv))
		SetMetric(ch, c.bytesOut, prometheus.GaugeValue, float64(netStat[0].BytesSent))
	} else {
		wzap.Info("[metric]", "error", err, "len(netStat)", len(netStat))
	}
	if fs, err := c.p.OpenFiles(); err == nil {
		SetMetric(ch, c.fds, prometheus.GaugeValue, float64(len(fs)))
	}

	SetMetric(ch, c.timeStamp, prometheus.GaugeValue, float64(time.Now().UnixNano()))
}
