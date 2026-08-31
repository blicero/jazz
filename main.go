// /home/krylon/go/src/github.com/blicero/jazz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-31 10:34:21 krylon>

package jazz

import (
	"fmt"

	"github.com/blicero/jazz/common"
)

func main() {
	fmt.Printf("%s %s\n%built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	fmt.Println("Nothing to see here, move along...")
}
