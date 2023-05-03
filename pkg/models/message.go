package models

import "net/http"

type Message struct {
	ID          string      `json:"id"`
	FlowID      string      `json:"flowID"`
	HttpHeaders http.Header `json:"httpHeaders"`
	RawQuery    string      `json:"rawQuery"`
	Payload     []byte      `json:"payload"`
}
