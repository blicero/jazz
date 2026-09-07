// /home/krylon/go/src/github.com/blicero/jazz/jes/01_jes_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-07 11:23:20 krylon>

package jes

import (
	"testing"
	"time"
)

var tj *JES

func TestJesCreate(t *testing.T) {
	var (
		err error
	)

	spoolDir = time.Now().Format("/tmp/jazz_jes_spool_20060102_150405")

	if tj, err = Create(spoolDir); err != nil {
		tj = nil
		t.Fatalf("Failed to create JES: %s\n",
			err.Error())
	}
} // func TestJesCreate(t *testing.T)
