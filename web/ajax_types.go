// /home/krylon/go/src/github.com/blicero/jazz/web/ajax_types.go
// -*- mode: go; coding: utf-8; -*-
// Created on 19. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-26 10:33:28 krylon>

package web

import "time"

// AjaxResponse is the regular content sent in response to Ajax or web service
// requests.
type AjaxResponse struct {
	Status    bool      `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}
