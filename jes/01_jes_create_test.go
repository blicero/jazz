// /home/krylon/go/src/github.com/blicero/jazz/jes/01_jes_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-07 10:47:45 krylon>

package jes

import (
	"os"
	"path/filepath"
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

func TestProcessJob(t *testing.T) {
	if tj == nil {
		t.SkipNow()
	}

	var (
		err     error
		fh      *os.File
		scripts []string
	)

	if fh, err = os.Open("testdata"); err != nil {
		t.Fatalf("Failed to open testdata: %s\n", err.Error())
	}

	defer fh.Close() // nolint: errcheck

	if scripts, err = fh.Readdirnames(-1); err != nil {
		t.Fatalf("Failed to read files from testdata: %s\n", err.Error())
	}

	for idx, script := range scripts {
		if !jclPat.MatchString(script) {
			continue
		} else if err = tj.processJob(filepath.Join("testdata", script)); err != nil {
			t.Fatalf("Failed to process script #%d %s: %s\n",
				idx,
				script,
				err.Error())
		}
	}
} // func TestProcessJob(t *testing.T)
