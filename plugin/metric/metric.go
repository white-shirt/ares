package metric

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sevenNt/wzap"
	"github.com/sevenNt/ares/server/echo"

	"github.com/rcrowley/go-metrics"
	"google.golang.org/grpc"
)

// Metric defines the config for Metric middleware.
type Metric struct {
	Options
	histograms map[string]metrics.Histogram
	keyLabels  map[string]map[string]string
	mu         sync.RWMutex
}

//New returns new metric instance.
func New(opts ...Option) *Metric {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	m := &Metric{
		Options:    options,
		histograms: make(map[string]metrics.Histogram),
		keyLabels:  make(map[string]map[string]string),
	}
	go m.watch()
	return m
}

// Func implements Middleware interface.
func (m *Metric) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			beg := time.Now()
			err = next(c)
			cost := time.Since(beg)
			m.update(nameFn(c.Request().Method+c.PatternPath()), cost, map[string]string{"if_type": "http"})
			m.update("http", cost, map[string]string{"if_type": "http"})
			return
		}
	}
}

func nameFn(method string) string {
	name := strings.Replace(method, "/", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, ":", "", -1)
	name = strings.Replace(name, "@", "", -1)
	name = strings.Replace(name, "#", "", -1)
	return name
}

// UnaryServerIntercept implements UnaryIntercept function of Interceptor.
func (m *Metric) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		beg := time.Now()
		resp, err = handler(ctx, req)
		cost := time.Since(beg)
		m.update("grpc", cost, map[string]string{"if_type": "grpc"})
		m.update(nameFn(info.FullMethod), cost, map[string]string{"if_type": "grpc"})
		return
	}
}

// StreamServerIntercept implements StreamIntercept function of Interceptor.
func (m *Metric) StreamServerIntercept() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		beg := time.Now()
		err = handler(srv, stream)
		cost := time.Since(beg)
		m.update(nameFn(info.FullMethod), cost, map[string]string{"if_type": "grpc"})
		m.update("grpc", cost, map[string]string{"if_type": "grpc"})
		return
	}
}

func (m *Metric) update(key string, duration time.Duration, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.histograms[key]; !ok {
		m.histograms[key] = metrics.NewRegisteredHistogram(key, nil, metrics.NewExpDecaySample(1028, 0.015))
	}
	m.histograms[key].Update(int64(duration))
	m.keyLabels[key] = labels
}

type (
	metric struct {
		key  string
		help string
		f    func() float64
	}
)

func (m *Metric) collect(key string, histogram metrics.Histogram) {
	wzap.Debugf("[metric] interface [%s],histogram[%d %f %f %f %f]",
		key, histogram.Count(),
		histogram.Percentile(0.99),
		histogram.Percentile(0.95),
		histogram.Percentile(0.90),
		histogram.Percentile(0.50))

	labels := m.keyLabels[key]
	for _, metric := range []metric{
		{"tps", "The request numbers of a specific URL", func() float64 { return float64(histogram.Count()) }},
		{"p99", "The p99 cost time of each interface", func() float64 { return histogram.Percentile(0.99) }},
		{"p95", "The p95 cost time of each interface", func() float64 { return histogram.Percentile(0.95) }},
		{"p90", "The p90 cost time of each interface", func() float64 { return histogram.Percentile(0.90) }},
		{"p50", "The p50 cost time of each interface", func() float64 { return histogram.Percentile(0.50) }},
	} {
		job := fmt.Sprintf("_%s_%s_%s", m.prefix, key, metric.key)
		mname := fmt.Sprintf("_%s_%s", m.prefix, metric.key)
		m.pusher.AddCollector(
			job,
			mname,
			metric.help,
			metric.f(),
			labels,
		)
	}
	histogram.Clear()
}

func (m *Metric) watch() {
	m.pusher.Start()
	ticker := time.Tick(m.interval)
	for {
		select {
		case <-ticker:
			m.mu.RLock()
			for key, histogram := range m.histograms {
				m.collect(key, histogram)
			}
			m.mu.RUnlock()
		}
	}
}

// Label implements plugin.Plugin interface
func (m *Metric) Label() string {
	return "metric"
}
