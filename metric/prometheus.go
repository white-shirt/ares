package metric

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"os"

	"github.com/sevenNt/wzap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

var (
	grouping map[string]string
)

const (
	//MaxMetricNum is the max number of channel to send metric
	MaxMetricNum = 1000
)

func init() {
	grouping = groupingName()
}

// NewPrometheusPusher constructs an instance of PrometheusPusher。
func NewPrometheusPusher(addr string) *PrometheusPusher {
	p := &PrometheusPusher{
		address: addr,
		metrics: make(chan *EchoMetricCollector, MaxMetricNum),
	}
	go p.sendLoop()
	return p
}

// EchoMetricCollector collect the metric from echo framework
type EchoMetricCollector struct {
	collector prometheus.Collector
	pushURL   string
}

// PrometheusPusher to send tps/p99 metric from echo framework
type PrometheusPusher struct {
	address string
	metrics chan *EchoMetricCollector
}

// AddCollector currently only support gauge type
func (m *PrometheusPusher) AddCollector(job, name, help string, value float64, labels ...prometheus.Labels) {
	gaugeLabels := make(map[string]string)
	for _, c := range labels {
		for k, v := range c {
			gaugeLabels[k] = v
		}
	}

	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        name,
		Help:        help,
		ConstLabels: gaugeLabels,
	})

	pushURL, err := parsePushURL(job, m.address, grouping)
	if err != nil {
		log.Warn("[metric]", "error", err)
		return
	}
	g.Set(value)

	m.metrics <- &EchoMetricCollector{
		collector: g,
		pushURL:   pushURL,
	}
}

// Start start prometheus pusher
func (m *PrometheusPusher) Start() error {
	go m.sendLoop()
	return nil
}

// sendLoop consume the golang metric channel
func (m *PrometheusPusher) sendLoop() {
	for c := range m.metrics {
		r := prometheus.NewRegistry()
		err := r.Register(c.collector)
		if err != nil {
			log.Warn("[metric]", "error ", err)
		}
		err = sendToAgent(r, c.pushURL, grouping)
	}
}

// AppMetricPusher the implement of application metric sender
type AppMetricPusher struct {
	name    string
	address string
	gather  prometheus.Gatherer
}

// Push implement sending the application metrics
func (m *AppMetricPusher) Push() error {
	pushURL, err := parsePushURL(m.name, m.address, grouping)
	if err != nil {
		return err
	}
	err = sendToAgent(m.gather, pushURL, grouping)
	return err
}

// groupingName to get the mapping of host name and instance
func groupingName() map[string]string {
	var groupingKey map[string]string
	hostname, err := os.Hostname()
	if err != nil {
		groupingKey = map[string]string{"instance": "unknown"}
	}
	groupingKey = map[string]string{"instance": hostname}
	return groupingKey
}

// parsePushURL to get the url sending http request
func parsePushURL(job, pushURL string, grouping map[string]string) (string, error) {
	if !strings.Contains(pushURL, "://") {
		pushURL = "http://" + pushURL
	}
	if strings.HasSuffix(pushURL, "/") {
		pushURL = pushURL[:len(pushURL)-1]
	}

	if strings.Contains(job, "/") {
		return "", fmt.Errorf("job contains '/': %s", job)
	}
	urlComponents := []string{url.QueryEscape(job)}
	for ln, lv := range grouping {
		if !model.LabelName(ln).IsValid() {
			return "", fmt.Errorf("grouping label has invalid name: %s", ln)
		}
		if strings.Contains(lv, "/") {
			return "", fmt.Errorf("value of grouping label %s contains '/': %s", ln, lv)
		}
		urlComponents = append(urlComponents, ln, lv)
	}
	pushURL = fmt.Sprintf("%s/metrics/job/%s", pushURL, strings.Join(urlComponents, "/"))
	return pushURL, nil
}

// sendToAgent send all metrics to monitor-agent
func sendToAgent(gather prometheus.Gatherer, pushURL string, grouping map[string]string) error {
	mfs, err := gather.Gather()
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtProtoDelim)

	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "job" {
					return fmt.Errorf("pushed metric %s (%s) already contains a job label", mf.GetName(), m)
				}
				if _, ok := grouping[l.GetName()]; ok {
					return fmt.Errorf(
						"pushed metric %s (%s) already contains grouping label %s",
						mf.GetName(), m, l.GetName(),
					)
				}
			}
		}
		enc.Encode(mf)
	}
	req, err := http.NewRequest(http.MethodPut, pushURL, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", string(expfmt.FmtProtoDelim))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		body, _ := ioutil.ReadAll(resp.Body) // Ignore any further error as this is for an error message only.
		return fmt.Errorf("unexpected status code %d while pushing to %s: %s", resp.StatusCode, pushURL, body)
	}
	return nil
}
