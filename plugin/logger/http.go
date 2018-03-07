package logger

import (
	"encoding/json"
)

type HTTPAccessMessage struct {
	Beg    int64           `json:"ts"`
	Cost   float64         `json:"cost"`
	Method string          `json:"method"`
	Status int             `json:"status"`
	Path   string          `json:"path"`
	Ext    json.RawMessage `json:"ext,omitempty"`
	Client string          `json:"client"`
}

//func (m *HTTPAccessMessage) MarshalJSON() ([]byte, error) {
//m.BegFormat = m.beg.Format("2006/01/02-15:04:05")
//return json.Marshal(m)
//}
