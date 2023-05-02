package models

import "net/http"

type Message struct {
	ID          string
	FlowID      string
	HttpHeaders http.Header
	Payload     []byte
}
