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
	"strconv"
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
	listenAddr   net.Addr
	tlsConfig    *tls.Config
	validRequest func(*http.Request) bool
	config       *Config
}

// loggerAdapter is a simple adapter around CoreDNS logger made to implement io.Writer in order to log errors from HTTP server
type loggerAdapter struct {
}

func (l *loggerAdapter) Write(p []byte) (n int, err error) {
	logger.Debugf("Writing HTTP request=%v", string(p))
	return len(p), nil
}

// NewServerHTTPS returns a new HTTPS server capable of performing DoH with dns
func NewServerHTTPS(addr string, dns dns.Handler, config *Config) (*ServerHTTPS, error) {
	var tlsConfig = config.DnsOverHttpServer.parsedTls

	// http/2 is recommended when using DoH. We need to specify it in next protos
	// or the upgrade won't happen.
	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	// Use a custom request validation func or use the standard DoH path check.

	srv := &http.Server{
		ReadTimeout:  time.Duration(config.DnsOverHttpServer.TimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(config.DnsOverHttpServer.TimeoutMs) * time.Millisecond,
		ErrorLog:     stdlog.New(&loggerAdapter{}, "", 0),
		Addr:         addr,
	}
	sh := &ServerHTTPS{
		handler: dns, tlsConfig: tlsConfig, httpsServer: srv, config: config,
	}

	srv.Handler = sh

	return sh, nil
}

// Stop stops the server. It blocks until the server is totally stopped.
func (s *ServerHTTPS) Stop() error {
	if s.httpsServer != nil {
		_ = s.httpsServer.Shutdown(context.Background())
	}
	return nil
}

// ServeHTTP is the handler that gets the HTTP request and converts to the dns format, calls the plugin
// chain, converts it back and write it to the client.
func (s *ServerHTTPS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !(r.URL.Path == pathDOH) {
		http.Error(w, "", http.StatusNotFound)
		s.countResponse(http.StatusNotFound)
		return
	}

	msg, err := requestToMsg(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		s.countResponse(http.StatusBadRequest)
		return
	}

	var writer = DohResponseWriter{remoteAddr: r.RemoteAddr}
	s.handler.ServeDNS(&writer, msg)

	// See section 4.2.1 of RFC 8484.
	// We are using code 500 to indicate an unexpected situation when the chain
	// handler has not provided any response message.
	if writer.msg == nil {
		http.Error(w, "No response", http.StatusInternalServerError)
		s.countResponse(http.StatusInternalServerError)
		return
	}

	buf, _ := writer.msg.Pack()

	age := s.config.TTL // seconds

	w.Header().Set("Content-Type", mimeTypeDOH)
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%v", age))
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.WriteHeader(http.StatusOK)
	s.countResponse(http.StatusOK)

	_, _ = w.Write(buf)
}

func (s *ServerHTTPS) countResponse(status int) {
	metric.DohResponseCount.With(prometheus.Labels{"status": fmt.Sprint(status)})
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
	buf, err := b64Enc.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	m := new(dns.Msg)
	err = m.Unpack(buf)

	return m, err
}

var b64Enc = base64.RawURLEncoding

var _ dns.ResponseWriter = &DohResponseWriter{}

type DohResponseWriter struct {
	msg        *dns.Msg
	remoteAddr string
}

func (w *DohResponseWriter) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) WriteMsg(msg *dns.Msg) error {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) Write(bytes []byte) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) TsigStatus() error {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) TsigTimersOnly(b bool) {
	//TODO implement me
	panic("implement me")
}

func (w *DohResponseWriter) Hijack() {
	//TODO implement me
	panic("implement me")
}
