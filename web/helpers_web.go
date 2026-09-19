// /home/krylon/go/src/github.com/blicero/jazz/web/helpers_web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 19. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-19 12:41:04 krylon>

package web

import (
	"encoding/json"
	"fmt"
)

func errJSON(msg string) []byte {
	var res = fmt.Sprintf(`{ "status": false, "message": %q }`,
		jsonEscape(msg))

	return []byte(res)
} // func errJSON(msg string) []byte

func jsonEscape(i string) string { // nolint: unused
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
} // func jsonEscape(i string) string
