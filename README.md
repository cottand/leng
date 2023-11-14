# leng

[![Go Report Card](https://goreportcard.com/badge/github.com/cottand/leng?style=flat-square)](https://goreportcard.com/report/github.com/cottand/leng)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/cottand/leng)
[![Build](https://github.com/Cottand/leng/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/Cottand/leng/actions/workflows/build.yml)

:zap: fast dns server, built to block advertisements and malware servers

Forked from [looterz/grimd](https://github.com/looterz/grimd)

# Features

- [x] DNS over UTP
- [x] DNS over TCP
- [x] DNS over HTTP(S) (DoH as per [RFC-8484](https://datatracker.ietf.org/doc/html/rfc8484))
- [x] Prometheus metrics API
- [x] Custom DNS records support
- [x] Blocklist fetching
- [x] Hardcoded blocklist config
- [x] Hardcoded whitelist config
- [x] Fast startup and tiny image _(so it can be quickly redeployed for use with templating for service discovery)_
- [x] Small memory footprint (~50MBs with metrics and DoH enabled)

# Installation

```
go install github.com/cottand/leng@latest
```

You can also 
- download one of the binary [releases](https://github.com/cottand/leng/releases)
- use the [Docker image](https://github.com/cottand/leng/pkgs/container/leng)
    - `docker run -d -p 53:53/udp -p 53:53/tcp -p 8080:8080/tcp ghcr.io/cottand/leng`
- use [Docker compose YML](https://raw.githubusercontent.com/cottand/leng/master/docker-compose.yml)
- use the [Nix flake](https://github.com/Cottand/leng/tree/master/flake.nix)
    - `nix run github:cottand/leng`

Detailed guides and resources can be found in the [docs website](https://cottand.github.com/leng).

# Configuration

By default, leng binds DNS to `0.0.0.0:53` and loads a few known blocklists. The default settings should be enough for
most.
See [the wiki](https://cottand.github.io/leng/Configuration) for the full config, including defaults and dynamic
config reloading.

### CLI Flags

```bash
$ leng -help

Usage of leng:
  -config string
    	location of the config file (default "leng.toml")
  -update
    	force an update of the blocklist database

```

# Building

Requires golang 1.21 or higher, you build leng like any other golang application, for example to build for linux x64

```shell
env GOOS=linux GOARCH=amd64 go build -v github.com/cottand/leng
```

# Building Docker

Run container and test

```shell
mkdir sources
docker build -t leng:latest -f docker/alpine.Dockerfile . && \
docker run -v $PWD/sources:/sources --rm -it -P --name leng-test leng:latest --config /sources/leng.toml --update
```

By default, if the program runs in a docker, it will automatically replace `127.0.0.1` in the default configuration
with `0.0.0.0` to ensure that the API interface is available.

```shell
curl -H "Accept: application/json" http://127.0.0.1:55006/application/active
```

# Objectives

- [x] ~~ARM64 Docker builds~~
- [ ] Better custom DNS support
    - [x] ~~Dynamic config reload for custom DNS issue#16~~
    - [x] ~~Fix multi-record responses issue#5~~
    - [x] ~~DNS record CNAME following issue#1~~
    - [ ] DNS record CNAME flattening a la cloudflare issue#27
    - [ ] Service discovery integrations? issue#4
- [x] Prometheus metrics exporter issue#3
- [x] DNS over HTTPS #2
- [ ] Add lots of docs

## Non-objectives

**Not keeping it simple**: I would like leng to become
a reliable custom DNS provider (like CoreDNS) and a reliable
adblocker (like Blocky) that has the perfect set of features
for self-hosters, and potentially for more critical setups.
