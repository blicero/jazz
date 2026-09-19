// /home/krylon/go/src/github.com/blicero/jazz/web/ajax_types.go
// -*- mode: go; coding: utf-8; -*-
// Created on 19. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-19 12:26:58 krylon>

package web

import "time"

type ajaxResponse struct {
	Status    bool      `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}
