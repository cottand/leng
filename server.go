package main

import (
	"github.com/cottand/leng/internal/metric"
	"time"

	"github.com/miekg/dns"
)

// Server type
type Server struct {
	host        string
	rTimeout    time.Duration
	wTimeout    time.Duration
	eventLoop   *EventLoop
	udpServer   *dns.Server
	tcpServer   *dns.Server
	httpServer  *ServerHTTPS
	udpHandler  *dns.ServeMux
	tcpHandler  *dns.ServeMux
	httpHandler *dns.ServeMux
}

// Run starts the server
func (s *Server) Run(config *Config, blockCache *MemoryBlockCache) {

	s.eventLoop = NewEventLoop(config, blockCache)

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.eventLoop.DoTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.eventLoop.DoUDP)

	httpHandler := dns.NewServeMux()
	httpHandler.HandleFunc(".", s.eventLoop.DoHTTP)

	s.tcpHandler = tcpHandler
	s.udpHandler = udpHandler
	s.httpHandler = httpHandler

	s.tcpServer = &dns.Server{
		Addr:         s.host,
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout,
	}

	s.udpServer = &dns.Server{
		Addr:         s.host,
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout,
	}

	if config.DnsOverHttpServer.Enabled {
		var err error
		timeout := time.Duration(config.DnsOverHttpServer.TimeoutMs) * time.Millisecond
		ttl := time.Duration(config.TTL) * time.Second
		s.httpServer, err = NewServerHTTPS(httpHandler, config.DnsOverHttpServer.Bind, timeout, ttl, config.DnsOverHttpServer.parsedTls)
		if err != nil {
			logger.Criticalf("failed to create http server %v", err)
		}
		go s.startHttp(config.DnsOverHttpServer.Bind)
	}
	go s.start(s.udpServer)
	go s.start(s.tcpServer)
}

func (s *Server) start(ds *dns.Server) {
	logger.Criticalf("start %s listener on %s\n", ds.Net, s.host)

	if err := ds.ListenAndServe(); err != nil {
		logger.Criticalf("start %s listener on %s failed: %s\n", ds.Net, s.host, err.Error())
	}
}

func (s *Server) startHttp(addr string) {
	logger.Criticalf("start http listener on %s\n", addr)

	if err := s.httpServer.ListenAndServe(); err != nil {
		logger.Criticalf("start http listener on %s failed or was closed: %s\n", addr, err.Error())
	}
}

// Stop stops the server
func (s *Server) Stop() {
	if s.eventLoop != nil {
		s.eventLoop.muActive.Lock()
		s.eventLoop.active = false
		close(s.eventLoop.requestChannel)
		s.eventLoop.muActive.Unlock()
	}
	if s.udpServer != nil {
		err := s.udpServer.Shutdown()
		if err != nil {
			logger.Critical(err)
		}
	}
	if s.tcpServer != nil {
		err := s.tcpServer.Shutdown()
		if err != nil {
			logger.Critical(err)
		}
	}

	if s.httpServer != nil {
		err := s.httpServer.Stop()
		if err != nil {
			logger.Critical(err)
		}
	}
}

// ReloadConfig only supports reloading the customDnsRecords section of the config for now
func (s *Server) ReloadConfig(config *Config) {
	newRecords := NewCustomDNSRecordsFromText(config.CustomDNSRecords)
	s.eventLoop.customDns = NewCustomRecordsResolver(newRecords)
	metric.CustomDNSConfigReload.Inc()
}
