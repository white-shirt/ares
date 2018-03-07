package config

type Observisor func(raw []byte, obj interface{})

// 订阅某个key的变化
func Subscribe(key string, observisor Observisor) {
}
