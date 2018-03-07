package metric

import (
	"time"

	"github.com/sevenNt/wzap"
)

// AppMetric ...
type AppMetric struct {
	pusher   *AppPusher
	interval time.Duration
	stop     chan struct{}
}

// NewAppMetric ...
func NewAppMetric(interval time.Duration, address, appID string, labels map[string]string) *AppMetric {
	m := &AppMetric{
		interval: interval,
		stop:     make(chan struct{}),
	}

	p := NewAppPusher(address, appID, labels)
	if p == nil {
		return nil
	}
	m.pusher = p
	go m.start()
	return m
}

// Stop ...
func (m *AppMetric) Stop() {
	close(m.stop)
}

func (m *AppMetric) start() {
	ticker := time.Tick(m.interval)
	for {
		select {
		case <-ticker:
			err := m.pusher.Push()
			if err != nil {
				wzap.Debug("[metric]", "error", err)
			}
		case <-m.stop:
			return
		}
	}
}
