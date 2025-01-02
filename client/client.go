package client

import (
	"github.com/DCSO/mauerspecht"
	"github.com/DCSO/mauerspecht/crypto"

	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	*http.Client
	cryptoCtx *crypto.Context
	url.URL
	id mauerspecht.ClientId
	mauerspecht.Config
	logentries   []mauerspecht.LogEntry
	ServerPubKey [32]byte
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("X-Specht-Id", c.id.String())
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("expected status code 200, got %d", res.StatusCode)
	}
	return res, nil
}

// dummyReadCloser is a stupid ReadCloser wrapper around simple io.Reader objects
type dummyReadCloser struct{ io.Reader }

func (*dummyReadCloser) Close() error { return nil }

func New(baseurl string, proxyurl string) (*Client, error) {
	var err error
	c := &Client{}
	c.Client = new(http.Client)
	*c.Client = *http.DefaultClient
	c.Client.Transport = new(http.Transport)
	*(c.Client.Transport.(*http.Transport)) = *(http.DefaultTransport.(*http.Transport))
	if proxyurl == "" {
		c.Client.Transport.(*http.Transport).Proxy = http.ProxyFromEnvironment
	} else {
		c.logf("Using proxy: %s", proxyurl)
		p, err := url.Parse(proxyurl)
		if err != nil {
			return nil, err
		}
		c.Client.Transport.(*http.Transport).Proxy = http.ProxyURL(p)
	}
	if ctx, err := crypto.NewContext(); err != nil {
		return nil, err
	} else {
		c.cryptoCtx = ctx
	}
	u, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}
	c.URL = *u
	if _, err := io.ReadFull(rand.Reader, c.id[:]); err != nil {
		return nil, err
	}
	c.logf("Client id: %s", c.id)
	c.log("Performing key exchange")
	u.Path = "/v1/kex"
	r, err := c.do(&http.Request{
		Method: "POST",
		URL:    u,
		Body:   &dummyReadCloser{bytes.NewReader(c.cryptoCtx.PubKey[:])},
	})
	if err != nil {
		return nil, err
	}
	buf, _ := io.ReadAll(r.Body)
	r.Body.Close()
	if len(buf) != 32 {
		return nil, fmt.Errorf("expected 32 byte key lengths from server, got %d", len(buf))
	}
	copy(c.ServerPubKey[:], buf)
	c.logf("Fetching configuration from %s", baseurl)
	u.Path = "/v1/config"
	if r, err = c.do(&http.Request{Method: "GET", URL: u}); err != nil {
		return nil, err
	}
	buf, err = io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return nil, err
	}
	if buf, err = c.cryptoCtx.Decrypt(&c.ServerPubKey, buf); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(buf, &c.Config); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) log(msg string) {
	log.Print(msg)
	c.logentries = append(c.logentries, mauerspecht.LogEntry{time.Now(), msg})
}

func (c *Client) logf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
	c.logentries = append(c.logentries, mauerspecht.LogEntry{time.Now(), fmt.Sprintf(msg, args...)})
}

type xfertype int

const (
	xferHeader xfertype = 1 << iota
	xferCookie
	xferBody
	xferMax
)

func (t xfertype) String() string {
	s := "xfer< "
	if t&xferHeader != 0 {
		s += "header "
	}
	if t&xferCookie != 0 {
		s += "cookie "
	}
	if t&xferBody != 0 {
		s += "body "
	}
	s += ">"
	return s
}

func (c *Client) download() {
	c.log("Trying downloads")
	for id := 0; id < len(c.Config.MagicStrings); id++ {
		for xfer := xfertype(1); xfer < xferMax; xfer++ {
			u := new(url.URL)
			*u = c.URL
			u.Path = "/v1/data"
			v := make(url.Values)
			v.Set("id", strconv.Itoa(id))
			if xfer&xferHeader != 0 {
				v.Set("header", "1")
			}
			if xfer&xferCookie != 0 {
				v.Set("cookie", "1")
			}
			if xfer&xferBody != 0 {
				v.Set("body", "1")
			}
			u.RawQuery = v.Encode()
			req := &http.Request{
				Method: "GET",
				URL:    u,
			}
			res, err := c.do(req)
			if err != nil {
				c.logf("download %s (%d): %v", xfer, id, err)
				continue
			}
			if xfer&xferHeader != 0 {
				if res.Header.Get("X-Specht") == c.Config.MagicStrings[id] {
					c.logf("download %s (%d): found string in header", xfer, id)
				} else {
					c.logf("download %s (%d): did not find string in header", xfer, id)
				}
			}
			if xfer&xferCookie != 0 {
				if res.Header.Get("Set-Cookie") == "Specht="+c.Config.MagicStrings[id] {
					c.logf("download %s (%d): found string in cookie", xfer, id)
				} else {
					c.logf("download %s (%d): did not find string in cookie", xfer, id)
				}
			}
			if xfer&xferBody != 0 {
				buf, _ := io.ReadAll(res.Body)
				if string(buf) == c.Config.MagicStrings[id] {
					c.logf("download %s (%d): found string in body", xfer, id)
				} else {
					c.logf("download %s (%d): did not find string in body", xfer, id)
				}
			}
		}
	}
}

func (c *Client) upload() {
	c.log("Trying uploads")
	for id := 0; id < len(c.Config.MagicStrings); id++ {
		for xfer := xfertype(1); xfer <= xferMax; xfer++ {
			u := c.URL
			u.Path = "/v1/data"
			req := &http.Request{
				Method: "POST",
				URL:    &u,
				Header: make(http.Header),
			}
			if xfer&xferHeader != 0 {
				req.Header.Set("X-Specht", c.Config.MagicStrings[id])
			}
			if xfer&xferCookie != 0 {
				req.Header.Set("Cookie", "Specht="+c.Config.MagicStrings[id])
			}
			if xfer&xferBody != 0 {
				req.Body = &dummyReadCloser{strings.NewReader(c.Config.MagicStrings[id])}
			}
			res, err := c.do(req)
			if err != nil {
				c.logf("upload %s (%d): %v", xfer, id, err)
				continue
			}
			buf, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				c.logf("upload %s (%d): %v", xfer, id, err)
				continue
			}
			buf, err = c.cryptoCtx.Decrypt(&c.ServerPubKey, buf)
			if err != nil {
				c.logf("upload %s (%d): %v", xfer, id, err)
				continue
			}
			response := mauerspecht.Response{-1, -1, -1}
			if err := json.Unmarshal(buf, &response); err != nil {
				c.logf("upload %s (%d): %v", xfer, id, err)
				continue
			}
			if xfer&xferHeader != 0 {
				if response.Header == id {
					c.logf("upload %s (%d): found string in header", xfer, id)
				} else {
					c.logf("upload %s (%d): did not find string in header", xfer, id)
				}
			}
			if xfer&xferCookie != 0 {
				if response.Cookie == id {
					c.logf("upload %s (%d): found string in cookie", xfer, id)
				} else {
					c.logf("upload %s (%d): did not find string in cookie", xfer, id)
				}
			}
			if xfer&xferBody != 0 {
				if response.Body == id {
					c.logf("upload %s (%d): found string in body", xfer, id)
				} else {
					c.logf("upload %s (%d): did not find string in body", xfer, id)
				}
			}
		}
	}
}

func (c *Client) report() {
	c.log("Submitting results to server")
	buf, err := json.Marshal(c.logentries)
	if err != nil {
		c.logf("Failed to marshal log messages: %v", err)
		return
	}
	buf, err = c.cryptoCtx.Encrypt(&c.ServerPubKey, buf)
	if err != nil {
		c.logf("Failed to encrypt message: %v", err)
		return
	}
	u := new(url.URL)
	*u = c.URL
	u.Path = "/v1/log"
	_, err = c.do(&http.Request{
		Method: "POST",
		URL:    u,
		Body:   &dummyReadCloser{bytes.NewReader(buf)},
	})
	if err != nil {
		c.logf("Failed to send log message: %v", err)
	}
}

func (c *Client) Run() {
	c.download()
	c.upload()
	c.report()
}
