package server

import (
	"github.com/DCSO/mauerspecht"
	"github.com/DCSO/mauerspecht/crypto"

	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	mauerspecht.Config
	http.Server
	cryptoCtx *crypto.Context
	Listeners []net.Listener
	PubKeys   map[mauerspecht.ClientId]*[32]byte
}

func getClient(r *http.Request) (*mauerspecht.ClientId, error) {
	c := &mauerspecht.ClientId{}
	if err := c.Set(r.Header.Get("X-Specht-ID")); err != nil {
		return nil, err
	}
	return c, nil
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return
}

func badRequest(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusBadRequest)
	return
}

func (s *Server) kex(w http.ResponseWriter, r *http.Request) {
	id, err := getClient(r)
	if err != nil || r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	var pubkey [32]byte
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil || len(buf) != 32 {
		badRequest(w)
		return
	}
	copy(pubkey[:], buf)
	s.PubKeys[*id] = &pubkey
	w.Write(s.cryptoCtx.PubKey[:])
	return
}

func (s *Server) config(w http.ResponseWriter, r *http.Request) {
	id, err := getClient(r)
	if err != nil || r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	pubkey, ok := s.PubKeys[*id]
	if !ok {
		badRequest(w)
		return
	}
	msg, err := json.Marshal(s.Config)
	if err != nil {
		internalServerError(w)
		return
	}
	buf, err := s.cryptoCtx.Encrypt(pubkey, msg)
	if err != nil {
		internalServerError(w)
		return
	}
	w.Write(buf)
	return
}

func (s *Server) patternPost(w http.ResponseWriter, r *http.Request) {
	id, err := getClient(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	pubkey, ok := s.PubKeys[*id]
	if !ok {
		badRequest(w)
		return
	}
	response := mauerspecht.Response{-1, -1, -1}
	var c, b string
	if c = r.Header.Get("Cookie"); strings.HasPrefix(c, "Specht=") {
		c = c[7:]
	} else {
		c = ""
	}
	h := r.Header.Get("X-Specht")
	if bd, err := ioutil.ReadAll(r.Body); err == nil {
		b = string(bd)
	}
	for i, s := range s.Config.MagicStrings {
		if c == s {
			response.Cookie = i
		}
		if h == s {
			response.Header = i
		}
		if b == s {
			response.Body = i
		}
	}
	msg, err := json.Marshal(response)
	if err != nil {
		internalServerError(w)
		return
	}
	buf, err := s.cryptoCtx.Encrypt(pubkey, msg)
	if err != nil {
		internalServerError(w)
		return
	}
	w.Write(buf)
	return
}

func (s *Server) patternGet(w http.ResponseWriter, r *http.Request) {
	patternId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || patternId >= len(s.Config.MagicStrings) {
		log.Printf("patternGet: %v", err)
		badRequest(w)
		return
	}
	pattern := s.Config.MagicStrings[patternId]
	if r.URL.Query().Get("header") == "1" {
		w.Header().Add("X-Specht", pattern)
	}
	if r.URL.Query().Get("cookie") == "1" {
		w.Header().Add("Set-Cookie", "Specht="+pattern)
	}
	if r.URL.Query().Get("body") == "1" {
		io.WriteString(w, pattern)
	}
	return
}

func (s *Server) pattern(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.patternGet(w, r)
	} else if r.Method == "POST" {
		s.patternPost(w, r)
	} else {
		http.NotFound(w, r)
	}
	return
}

func (s *Server) log(w http.ResponseWriter, r *http.Request) {
	id, err := getClient(r)
	if err != nil || r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	pubkey, ok := s.PubKeys[*id]
	if !ok {
		badRequest(w)
		return
	}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(w)
		return
	}
	if buf, err = s.cryptoCtx.Decrypt(pubkey, buf); err != nil {
		badRequest(w)
		return
	}
	var logentries []mauerspecht.LogEntry
	if err = json.Unmarshal(buf, &logentries); err != nil {
		badRequest(w)
		return
	}
	for _, l := range logentries {
		log.Printf("%s %s %s", l.TS.Format(time.RFC3339), id, l.Msg)
	}
	return
}

func logo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(
		`
(*-> |             | <-*)
())| | Mauerspecht | |(()
 \"| | thcepsreuaM | |"/
  '| |             | |'

`))
}

func (s *Server) setupServer() {
	m := http.NewServeMux()
	m.HandleFunc("/", logo)
	m.HandleFunc("/v1/kex", s.kex)
	m.HandleFunc("/v1/config", s.config)
	m.HandleFunc("/v1/data", s.pattern)
	m.HandleFunc("/v1/log", s.log)
	s.Server = http.Server{Handler: m}
}

func New(config mauerspecht.Config) (*Server, error) {
	s := Server{
		Config:  config,
		PubKeys: make(map[mauerspecht.ClientId]*[32]byte),
	}
	if ctx, err := crypto.NewContext(); err != nil {
		return nil, err
	} else {
		s.cryptoCtx = ctx
	}
	if s.Config.Hostname == "" {
		s.Config.Hostname = "localhost"
	}
	if len(s.Config.MagicStrings) == 0 {
		s.Config.MagicStrings = []string{
			`X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`,
		}
	}
	if len(s.Config.HTTPPorts) == 0 {
		s.Config.HTTPPorts = []int{80}
	}
	s.setupServer()
	for _, port := range config.HTTPPorts {
		l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			log.Printf("listen: %d: %v", port, err)
			continue
		}
		s.Listeners = append(s.Listeners, l)
		go s.Serve(l)
	}
	return &s, nil
}

func (s *Server) Close() error {
	for _, l := range s.Listeners {
		l.Close()
	}
	s.Server.Close()
	s.Listeners = nil
	return nil
}
