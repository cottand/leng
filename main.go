package main

import (
	"context"
	"flag"
	"golang.org/x/sys/unix"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

var (
	configPath     string
	forceUpdate    bool
	lengActive     bool
	lengActivation *ActivationHandler
)

func reloadBlockCache(config *Config,
	blockCache *MemoryBlockCache,
	questionCache *MemoryQuestionCache,
	apiServer *http.Server,
	server *Server,
	reloadChan chan bool) (*MemoryBlockCache, *http.Server, error) {

	logger.Debug("Reloading the blockcache")
	blockCache = PerformUpdate(config, true)
	server.Stop()
	if apiServer != nil {
		if err := apiServer.Shutdown(context.Background()); err != nil {
			logger.Debugf("error shutting down api server: %v", err)
		}
	}
	server.Run(config, blockCache, questionCache)
	apiServer, err := StartAPIServer(config, reloadChan, blockCache, questionCache)
	if err != nil {
		logger.Fatal(err)
		return nil, nil, err
	}

	return blockCache, apiServer, nil
}

func main() {
	flag.Parse()

	config, err := LoadConfig(configPath)
	if err != nil {
		logger.Fatal(err)
	}

	loggingState, err := loggerInit(config.LogConfig)
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		loggingState.cleanUp()
	}()

	lengActive = true
	quitActivation := make(chan bool)
	actChannel := make(chan *ActivationHandler)

	go startActivation(actChannel, quitActivation)
	lengActivation = <-actChannel
	close(actChannel)

	server := &Server{
		host:     config.Bind,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}

	// BlockCache contains all blocked domains
	blockCache := &MemoryBlockCache{Backend: make(map[string]bool)}
	// QuestionCache contains all queries to the dns server
	questionCache := makeQuestionCache(config.QuestionCacheCap)

	reloadChan := make(chan bool)

	// The server will start with an empty blockcache soe we can dowload the lists if leng is the
	// system's dns server.
	server.Run(config, blockCache, questionCache)

	var apiServer *http.Server
	// Load the block cache, restart the server with the new context
	blockCache, apiServer, err = reloadBlockCache(config, blockCache, questionCache, apiServer, server, reloadChan)

	if err != nil {
		logger.Fatalf("Cannot start the API server %s", err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, unix.SIGHUP, unix.SIGUSR1)

forever:
	for {
		select {
		case s := <-sig:
			switch s {
			case os.Interrupt:
				logger.Error("SIGINT received, stopping\n")
				quitActivation <- true
				break forever
			case unix.SIGHUP:
				logger.Error("SIGHUP received: rotating logs\n")
				err := loggingState.reopen()
				if err != nil {
					logger.Error(err)
				}
			case unix.SIGUSR1:
				logger.Info("SIGUSR1 received: reloading config\n")
				reloadConfigFromFile(server)
			}
		case <-reloadChan:
			blockCache, apiServer, err = reloadBlockCache(config, blockCache, questionCache, apiServer, server, reloadChan)
			if err != nil {
				logger.Fatalf("Cannot start the API server %s", err)
			}
		}
	}
	// make sure we give the activation goroutine time to exit
	<-quitActivation
	logger.Debugf("Main goroutine exiting")
}

func init() {
	flag.StringVar(&configPath, "config", "leng.toml", "location of the config file")
	flag.BoolVar(&forceUpdate, "update", false, "force an update of the blocklist database")

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func reloadConfigFromFile(s *Server) {
	config, err := LoadConfig(configPath)
	if err != nil {
		logger.Errorf("Failed to reload config %v", err)
	}
	s.ReloadConfig(config)
}
