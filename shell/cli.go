// /home/krylon/go/src/github.com/blicero/jazz/shell/cli.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-08 11:01:29 krylon>

package shell

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Feralthedogg/go-functional/pkg/functional"
	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	prompt "github.com/c-bata/go-prompt"
)

const pprompt = "~> "

// Shell is a simple command line shell that prompts for commands to execute
// as a Job.
type Shell struct {
	log   *log.Logger
	shell *prompt.Prompt
	j     *model.Job
}

// Create creates (well, duh) and returns a fresh Shell.
func Create() (*Shell, error) {
	var (
		err error
		env []string
		s   = &Shell{
			j: &model.Job{
				Steps: make([]model.Step, 0),
				Env:   make(map[string]string),
			},
		}
	)

	if s.log, err = common.GetLogger(logdomain.Shell); err != nil {
		return nil, err
	} else if s.j.WorkDir, err = os.Getwd(); err != nil {
		return nil, err
	}

	env = os.Environ()

	for _, v := range env {
		var (
			pieces    []string
			name, val string
		)

		pieces = strings.Split(v, "=")
		name = pieces[0]
		val = pieces[1]

		s.j.Env[name] = val
	}

	s.shell = prompt.New(
		s.executor,
		s.completerExecutables,
		prompt.OptionPrefix(pprompt))

	return s, nil
} // func Create() (*Shell, error)

// Run runs the Shell. If everything goes well, it returns a Job that can be
// submitted to the monitor.
func (s *Shell) Run() (j *model.Job, e error) {

	defer func() {
		if x := recover(); x != nil {
			s.log.Printf("[ERROR] Panic while prompting for Job: %#v\n",
				x)
			e = x.(error)
		}
	}()

	s.shell.Run()

	return s.j, nil
}

func (s *Shell) executor(input string) {
	s.log.Printf("[TRACE] Executing the following command: %s\n",
		input)

} // func (s *Shell) executor(input string

// completerExecutables looks up all executable files in the folders from PATH
// and returns the ones that match the text typed so far.
func (s *Shell) completerExecutables(d prompt.Document) []prompt.Suggest {
	var (
		err                      error
		sugg                     = []prompt.Suggest{}
		path, word               string
		executables, pathFolders []string
	)

	if path = os.Getenv("PATH"); path == "" {
		s.log.Printf("[ERROR] Environment variable PATH is empty\n")
		return sugg
	}

	pathFolders = strings.Split(path, ":")

	for _, dir := range pathFolders {
		var files []string

		if files, err = s.readFolder(dir); err != nil {
			s.log.Printf("[ERROR] Cannot read %s: %s\n",
				dir,
				err.Error())
			return sugg
		}

		executables = append(executables, files...)
	}

	word = d.GetWordAfterCursor()

	for _, ex := range executables {
		var base = filepath.Base(ex)
		if strings.HasPrefix(base, word) {
			s := prompt.Suggest{
				Text:        base,
				Description: ex,
			}

			sugg = append(sugg, s)
		}
	}

	return sugg
} // func (s *Shell) completerExecutables(d prompt.Document) []prompt.Suggest

func (s *Shell) readFolder(path string) ([]string, error) {
	var (
		err   error
		fh    *os.File
		info  []os.FileInfo
		files []string
	)

	if fh, err = os.Open(path); err != nil {
		return nil, err
	}

	defer fh.Close() // nolint: errcheck

	if info, err = fh.Readdir(-1); err != nil {
		return nil, err
	}

	info = slices.DeleteFunc(info, func(i os.FileInfo) bool {
		return i.Mode().IsRegular() &&
			i.Mode()&0111 != 0
	})

	files = functional.Map(func(i os.FileInfo) string {
		return filepath.Join(path, i.Name())
	}, info)

	return files, nil
} // func (s *Shell) readFolder(path string) ([]string, error)
