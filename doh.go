package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/cottand/grimd/internal/metric"
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"time"
)

/**
This implementation is heavily inspired by CoreDNS and used as per their Apache 2 license
see https://github.com/coredns/coredns/blob/v1.11.1/core/dnsserver/server_https.go

There is no NOTICE redistribution as, at the time of producing the derivative work, CoreDNS did
not distribute such a notice with their work.
*/

const mimeTypeDOH = "application/dns-message"

// pathDOH is the URL path that should be used.
const pathDOH = "/dns-query"

// ServerHTTPS represents an instance of a DNS-over-HTTPS server.
type ServerHTTPS struct {
	Net          string
	handler      dns.Handler
	httpsServer  *http.Server
	tlsConfig    *tls.Config
	validRequest func(*http.Request) bool
	bind         string
	ttl          time.Duration
}

// loggerAdapter is a simple adapter around CoreDNS logger made to implement io.Writer in order to log errors from HTTP server
type loggerAdapter struct {
}

func (l *loggerAdapter) Write(p []byte) (n int, err error) {
	logger.Debugf("Writing HTTP request=%v", string(p))
	return len(p), nil
}

// NewServerHTTPS returns a new HTTPS server capable of performing DoH with dns
func NewServerHTTPS(
	dns dns.Handler,
	bind string,
	timeout time.Duration,
	ttl time.Duration,
	tls *tls.Config,
) (*ServerHTTPS, error) {

	// http/2 is recommended when using DoH. We need to specify it in next protos
	// or the upgrade won't happen.
	if tls != nil {
		tls.NextProtos = []string{"h2", "http/1.1"}
	}

	// Use a custom request validation func or use the standard DoH path check.

	srv := &http.Server{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		ErrorLog:     stdlog.New(&loggerAdapter{}, "", 0),
		Addr:         bind,
	}
	sh := &ServerHTTPS{
		handler: dns, httpsServer: srv, ttl: ttl, bind: bind,
	}
	srv.Handler = sh

	return sh, nil
}

func (s *ServerHTTPS) ListenAndServe() error {
	return s.httpsServer.ListenAndServe()
}

// Stop stops the server. It blocks until the server is totally stopped.
func (s *ServerHTTPS) Stop() error {
	if s.httpsServer != nil {
		_ = s.httpsServer.Shutdown(context.Background())
	}
	return nil
}

// ServeHTTP is the eventLoop that gets the HTTP request and converts to the dns format, calls the resolver,
// converts it back and write it to the client.
func (s *ServerHTTPS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !(r.URL.Path == pathDOH) {
		http.Error(w, "", http.StatusNotFound)
		countResponse(http.StatusNotFound)
		return
	}

	msg, err := requestToMsg(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		countResponse(http.StatusBadRequest)
		logger.Noticef("error when serving DoH request: %v", err)
		return
	}

	var writer = DohResponseWriter{remoteAddr: r.RemoteAddr, host: r.Host, delegate: w, completed: make(chan empty, 1)}
	s.handler.ServeDNS(&writer, msg)
	_, ok := <-writer.completed
	if writer.err != nil || ok != true {
		return
	}

	age := s.ttl // seconds

	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%v", age.Seconds()))
}

func countResponse(status int) {
	metric.DohResponseCount.With(prometheus.Labels{"status": fmt.Sprint(status)}).Inc()
}

// Shutdown stops the server (non gracefully).
func (s *ServerHTTPS) Shutdown() {
	if s.httpsServer != nil {
		_ = s.httpsServer.Shutdown(context.Background())
	}
}

func requestToMsg(req *http.Request) (*dns.Msg, error) {
	if req.Method == "GET" {
		return getRequestToMsg(req)
	}
	if req.Method == "POST" {
		return postRequestToMsg(req)
	}
	return nil, fmt.Errorf("unexpected method for DoH request %v", req.Method)
}

// postRequestToMsg extracts the dns message from the request body.
func postRequestToMsg(req *http.Request) (*dns.Msg, error) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(req.Body)

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	m := new(dns.Msg)
	err = m.Unpack(buf)
	return m, err
}

// getRequestToMsg extract the dns message from the GET request.
func getRequestToMsg(req *http.Request) (*dns.Msg, error) {
	values := req.URL.Query()
	b64, ok := values["dns"]
	if !ok {
		return nil, fmt.Errorf("no 'dns' query parameter found")
	}
	if len(b64) != 1 {
		return nil, fmt.Errorf("multiple 'dns' query values found")
	}
	return base64ToMsg(b64[0])
}

func base64ToMsg(b64 string) (*dns.Msg, error) {
	buf, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	m := new(dns.Msg)
	err = m.Unpack(buf)

	return m, err
}

type empty struct{}

// DohResponseWriter implements dns.ResponseWriter
type DohResponseWriter struct {
	msg        *dns.Msg
	remoteAddr string
	delegate   http.ResponseWriter
	host       string
	err        error
	completed  chan empty
}

// See section 4.2.1 of RFC 8484.
// We are using code 500 to indicate an unexpected situation when the chain
// eventLoop has not provided any response message.
func (w *DohResponseWriter) handleErr(err error) {
	logger.Warningf("error when replying to DoH: %v", err)
	http.Error(w.delegate, "No response", http.StatusInternalServerError)
	countResponse(http.StatusInternalServerError)
	w.err = err
	return
}

func (w *DohResponseWriter) LocalAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", w.remoteAddr)
	return addr
}

func (w *DohResponseWriter) RemoteAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", w.remoteAddr)
	return addr
}

func (w *DohResponseWriter) WriteMsg(msg *dns.Msg) error {
	defer func() {
		w.completed <- empty{}
		close(w.completed)
	}()
	w.msg = msg
	buf, err := msg.Pack()
	if err != nil {
		w.handleErr(err)
		return err
	}
	w.delegate.Header().Set("Content-Type", mimeTypeDOH)
	_, err = w.Write(buf)
	if err != nil {
		w.handleErr(err)
		return err
	}
	countResponse(http.StatusOK)
	return nil
}

func (w *DohResponseWriter) Write(bytes []byte) (int, error) {
	return w.delegate.Write(bytes)
}

func (w *DohResponseWriter) Close() error {
	return nil
}

func (w *DohResponseWriter) TsigStatus() error {
	return nil
}

func (w *DohResponseWriter) TsigTimersOnly(_ bool) {
}

func (w *DohResponseWriter) Hijack() {
	return
}
