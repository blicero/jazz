// /home/krylon/go/src/github.com/blicero/jazz/model/wsresponse.go
// -*- mode: go; coding: utf-8; -*-
// Created on 26. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-01 10:51:44 krylon>

package model

import "time"

// WebResponse is the regular content sent in response to Ajax or web service
// requests.
type WebResponse struct {
	Status    bool      `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Payload   string    `json:"payload"`
}
