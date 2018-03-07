package logger

import (
	"encoding/json"
)

type GRPCAccessMessage struct {
	Beg       int64           `json:"ts"`
	Cost      float64         `json:"cost"`
	Method    string          `json:"method"`
	Ext       json.RawMessage `json:"ext,omitempty"`
	Client    string          `json:"client"`
	BegFormat string          `json:"ts"`
	Error     error           `json:"err"`
}

//func (m *GRPCAccessMessage) MarshalJSON() ([]byte, error) {
//m.BegFormat = m.beg.Format("2006/01/02-15:04:05")
//return json.Marshal(m)
//}
