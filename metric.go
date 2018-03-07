package ares

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/plugin/metric"
	"github.com/sevenNt/hera"
	"github.com/sevenNt/wzap"
)

func (app *App) initMetric() {
	metricAddr := hera.GetString("app.metric.addr")
	if metricAddr == "" || application.ID() == "" {
		wzap.Warnf("[metric]metric address is not setting, no metric will gather")
		return
	}

	//当前应用所有服务的键值对
	labels := make(map[string]string)
	labels["pid"] = strconv.Itoa(os.Getpid())
	labels["app_name"] = application.Name()
	for _, j := range app.servers {
		if v, ok := labels[j.Scheme()]; ok {
			labels[j.Scheme()] = fmt.Sprintf("%s_%s", v, j.Addr())
		} else {
			labels[j.Scheme()] = j.Addr()
		}
	}
	wzap.Infof("[metric]application servers list is %v", labels)

	interval := hera.GetDuration("metric.interval")
	if interval == 0 {
		interval = time.Second * 10
	}

	wzap.Infof("[metric]application metric sending to %q with interval %v", metricAddr, interval)
	m := metric.NewAppMetric(interval, metricAddr, application.ID(), labels)
	if m == nil {
		wzap.Warnf("[metric]fail to create application metric")
	}
}
