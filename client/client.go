// /home/krylon/go/src/github.com/blicero/jazz/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 10. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-02 11:02:35 krylon>

// Package client wraps the communication with the Job Queue web service.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/blicero/jazz/client/cmd"
	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
)

// MIME-Type for JSON
const mjson = "application/json"

var endpoints = map[cmd.Cmd]string{
	cmd.Submit:     "/ws/job/new",
	cmd.QueryQueue: "/ws/job/all",
}

// Client handles the communication with the Job Queue web service.
type Client struct {
	log    *log.Logger
	client http.Client
	addr   string
	uri    *url.URL
}

// New creates a fresh Client that will talk to the given address.
func New(addr string) (*Client, error) {
	if addr == "" {
		addr = fmt.Sprintf("[::]:%d", common.WebPort)
	}

	var (
		err  error
		ustr string
		c    = &Client{addr: addr}
	)

	if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	}

	ustr = fmt.Sprintf("http://%s/", addr)

	if c.uri, err = url.Parse(ustr); err != nil {
		c.log.Printf("[CRITICAL] Cannot parse web service URL %q: %s\n",
			ustr,
			err.Error())
		return nil, err
	}

	return c, nil
} // func New(addr string) (*Client, error)

// GetURL returns the URL for the given operation.
func (c *Client) GetURL(op cmd.Cmd) string {
	var u = c.uri.Clone()
	u.Path = endpoints[op]

	return u.String()
} // func (c *Client) GetURL(op cmd.Cmd) *url.URL

// JobSubmit submits a Job to the Queue. If successful, the Job instance
// has its ID filled in.
func (c *Client) JobSubmit(job *model.Job) error {
	var (
		err  error
		sbuf []byte
		res  *http.Response
		resp model.WebResponse
		buf  bytes.Buffer
		uri  string
	)

	if sbuf, err = json.Marshal(job); err != nil {
		c.log.Printf("[ERROR] Cannot serialize Job: %s\n",
			err.Error())
		return err
	}

	buf = *bytes.NewBuffer(sbuf)
	uri = c.GetURL(cmd.Submit)

	if res, err = c.client.Post(uri, mjson, &buf); err != nil {
		c.log.Printf("[ERROR] Failed to speak to web service @ %s: %s\n",
			uri,
			err.Error())
		return err
	} else if res == nil {
		err = fmt.Errorf("http.Post returned no error, but nil return response")
		c.log.Printf("[ERROR] %s\n", err.Error())
		return err
	}

	defer res.Body.Close() // nolint: errcheck

	switch res.StatusCode {
	case 200:
		if _, err = io.Copy(&buf, res.Body); err != nil {
			c.log.Printf("[ERROR] Cannot copy response Body: %s\n",
				err.Error())
			return err
		} else if err = json.Unmarshal(buf.Bytes(), &resp); err != nil {
			c.log.Printf("[ERROR] Cannot unmarshal response: %s\n%s\n\n",
				err.Error(),
				buf.String())
			return err
		}

		var id int64

		if id, err = strconv.ParseInt(resp.Message, 10, 64); err != nil {
			c.log.Printf("[ERROR] Cannot parse Job ID %q: %s\n",
				resp.Message,
				err.Error())
			return err
		}

		c.log.Printf("[INFO] Job %d submitted successfully\n", id)
		job.ID = id
	case 500:
		// We are in trouble
		err = fmt.Errorf("server-side error attempting to submit Job: %s",
			res.Status)
		return err
	default:
		c.log.Printf("[DEBUG] Unexpected HTTP status code from daemon: %s\n",
			res.Status)

	}

	return nil
} // func (c *Client) JobSubmit(job *model.Job) error

// QueryQueue asks for the Jobs currently enqueued.
func (c *Client) QueryQueue() ([]*model.Job, error) {
	var (
		err   error
		res   *http.Response
		reply model.WebResponse
		buf   bytes.Buffer
		uri   string
		jobs  []*model.Job
	)

	uri = c.GetURL(cmd.QueryQueue)

	if res, err = c.client.Get(uri); err != nil {
		c.log.Printf("[ERROR] Failed to send request %s to web service at %s: %s\n",
			cmd.QueryQueue,
			uri,
			err.Error())
		return nil, err
	}

	defer res.Body.Close() // nolint: errcheck
	jobs = make([]*model.Job, 0)

	switch res.StatusCode {
	case 200:
		if _, err = io.Copy(&buf, res.Body); err != nil {
			c.log.Printf("[ERROR] Failed to read HTTP response body from %s: %s\n",
				uri,
				err.Error())
			return nil, err
		} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
			c.log.Printf("[ERROR] Cannot parse response from %s: %s\n\n%q\n",
				uri,
				err.Error(),
				buf.String())
			return nil, err
		} else if !reply.Status {
			err = fmt.Errorf("web Service request failed: %s",
				reply.Message)
			c.log.Printf("[ERROR] %s\n", err.Error())
			return nil, err
		} else if err = json.Unmarshal([]byte(reply.Payload), &jobs); err != nil {
			c.log.Printf("[ERROR] Cannot parse Payload: %s\n\n%s\n\n",
				err.Error(),
				reply.Payload)
			return nil, err
		}

		return jobs, nil
	case 500:
		c.log.Printf("[ERROR] Server-side error %s\n",
			res.Status)
		return nil, fmt.Errorf("server-side error %s", res.Status)
	default:
		err = fmt.Errorf("don't know how to deal with HTTP response code %s",
			res.Status)
		c.log.Printf("[ERROR] %s\n", err.Error())
		return nil, err
	}
} // func (c *Client) QueryQueue() ([]*model.Job, error)
