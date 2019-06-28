package mauerspecht

import (
	"time"
)

type Response struct {
	Header int `json:"header"`
	Cookie int `json:"Cookie"`
	Body   int `json:"body"`
}

type LogEntry struct {
	TS  time.Time `json:"timestamp"`
	Msg string    `json:"message"`
}
