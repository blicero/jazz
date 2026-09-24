// /home/krylon/go/src/github.com/blicero/jazz/jes/01_jes_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-24 15:28:37 krylon>

package web

import (
	"fmt"
	"testing"

	"github.com/blicero/jazz/common"
)

var tsrv *Web

func TestCreateServer(t *testing.T) {
	var (
		err  error
		addr = fmt.Sprintf(":%d", common.WebPort+2)
	)

	if tsrv, err = Create(addr, nil); err != nil {
		tsrv = nil
		t.Fatalf("Cannot create Web server: %s",
			err.Error())
	}
} // func TestCreateServer(t *testing.T)
