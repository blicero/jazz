// /home/krylon/go/src/github.com/blicero/jazz/model/predicate/predicate.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-31 10:25:20 krylon>

//go:generate stringer -type=ID

// Package predicate defines constants to identify predicates for Job steps.
package predicate

// ID identifies a predicate.
type ID uint8

const (
	None ID = iota
	Always
	And
	Or
)
