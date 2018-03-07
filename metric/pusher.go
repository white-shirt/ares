package metric

// Pusher represents metric pusher
type Pusher interface {
	AddCollector(job, name, help string, value float64, labelMap map[string]string)
	Start() error
}
