# grimd
[![Go Report Card](https://goreportcard.com/badge/github.com/cottand/grimd?style=flat-square)](https://goreportcard.com/report/github.com/cottand/grimd)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/cottand/grimd)
[![Release](https://github.com/cottand/grimd/actions/workflows/release.yaml/badge.svg)](https://github.com/looterz/cottand/releases)

:zap: Fast dns proxy, built to black-hole internet advertisements and malware servers. Capable of custom DNS.

Forked from [looterz/grimd](https://github.com/looterz/grimd)

# Installation
```
go install github.com/cottand/grimd@latest
```

You can also download one of the [releases](https://github.com/cottand/grimd/releases) or [docker images](https://github.com/cottand/grimd/pkgs/container/grimd). Detailed guides and resources can be found on the [wiki](https://github.com/cottand/grimd/wiki).

# Docker Installation
To quickly get grimd up and running with docker, run
```
docker run -d -p 53:53/udp -p 53:53/tcp -p 8080:8080/tcp ghcr.io/cottand/grimd:latest
```

Alternatively, download the [docker-compose.yml](https://raw.githubusercontent.com/cottand/grimd/master/docker-compose.yml) file and launch it using docker-compose.
```
docker-compose up -d
```

# Configuration

See [the wiki](https://github.com/Cottand/grimd/wiki/Configuration)

# Building
Requires golang 1.7 or higher, you build grimd like any other golang application, for example to build for linux x64
```shell
env GOOS=linux GOARCH=amd64 go build -v github.com/looterz/grimd
```

# Building Docker
Run container and test
```shell
mkdir sources
docker build -t grimd:latest -f docker/alpine.Dockerfile . && \
docker run -v $PWD/sources:/sources --rm -it -P --name grimd-test grimd:latest --config /sources/grimd.toml --update
```

By default, if the program runs in a docker, it will automatically replace `127.0.0.1` in the default configuration with `0.0.0.0` to ensure that the API interface is available.

```shell
curl -H "Accept: application/json" http://127.0.0.1:55006/application/active
```

# Web API
A restful json api is exposed by default on the local interface, allowing you to build web applications that visualize requests, blocks and the cache. [reaper](https://github.com/looterz/reaper) is the default grimd web frontend.


If you want to enable the default dashboard, make sure the configuration file contains the following:

```toml
dashboard = true
```

![reaper-example](http://i.imgur.com/oXLtqSz.png)

# Speed
Incoming requests spawn a goroutine and are served concurrently, and the block cache resides in-memory to allow for rapid lookups, while answered queries are cached allowing grimd to serve thousands of queries at once while maintaining a memory footprint of under 15mb for 100,000 blocked domains!

# Daemonize
You can find examples of different daemon scripts for grimd on the [wiki](https://github.com/looterz/grimd/wiki/Daemon-Scripts).

# TODO - objectives 

These are some of the things I would like to contribute in this fork:
- [x] ~~ARM64 Docker builds~~
- [ ] Better custom DNS support (DNS flattening #1, multiple records #5 )
    - [ ] Service discovery integrations # 4
- [ ] Prometheus metrics exporter #3
- [ ] DNS over HTTPS #2
- [ ] Add lots of docs

## Non-objectives
**Not keeping it simple**: I would like grimd to become
a reliable custom DNS provider (like CoreDNS) and a reliable
adblocker (like Blocky) that has the perfect set of features
for self-hosters, and potentially for more critical setups.