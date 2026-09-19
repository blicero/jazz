// /home/krylon/go/src/github.com/blicero/jazz/jes/00_main_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-07 10:35:30 krylon>

package jes

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/blicero/jazz/common"
)

var spoolDir string

func TestMain(m *testing.M) {
	var (
		err     error
		result  int
		baseDir = time.Now().Format("/tmp/jazz_jes_test_20060102_150405")
	)

	if err = common.SetBaseDir(baseDir); err != nil {
		fmt.Printf("Cannot set base directory to %s: %s\n",
			baseDir,
			err.Error())
		os.Exit(1)
	} else if result = m.Run(); result == 0 {
		// If any test failed, we keep the test directory (and the
		// database inside it) around, so we can manually inspect it
		// if needed.
		// If all tests pass, OTOH, we can safely remove the directory.
		fmt.Printf("Removing BaseDir %s\n",
			baseDir)
		_ = os.RemoveAll(baseDir)
		_ = os.RemoveAll(spoolDir)
	} else {
		fmt.Printf(">>> TEST DIRECTORY: %s\n", baseDir)
	}

	os.Exit(result)
} // func TestMain(m *testing.M)
