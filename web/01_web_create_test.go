// /home/krylon/go/src/github.com/blicero/jazz/web/01_web_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-30 10:46:18 krylon>

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
