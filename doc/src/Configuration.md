If leng.toml is not found the default configuration will be used. If it is found, fields that are set will act as overrides.

## Quick Start

If you are happy to use Cloudflare as your upstream DNS provider and just want to generally block tracking
and advertising, the following minimal config should be enough.

If you want to tweak more settings, keep scrolling down!

```toml
# address to bind to for the DNS server
bind = "0.0.0.0:53"

# address to bind to for the API server
api = "127.0.0.1:8080"

# manual custom dns entries - comments for reference
customdnsrecords = [
    # "example.mywebsite.tld      IN A       10.0.0.1",
]

[Metrics]
    enabled = false

[Blocking]
    # manual whitelist entries - comments for reference
    whitelist = [
        # "getsentry.com",
    ]
```

## Default configuration

```toml
# log configuration
# format: comma separated list of options, where options is one of 
#   file:<filename>@<loglevel>
#   stderr>@<loglevel>
#   syslog@<loglevel>
# loglevel: 0 = errors and important operations, 1 = dns queries, 2 = debug
# e.g. logconfig = "file:leng.log@2,syslog@1,stderr@2"
logconfig = "stderr@2"

# apidebug enables the debug mode of the http api library
apidebug = false

# address to bind to for the DNS server
bind = "0.0.0.0:53"

# address to bind to for the API server
api = "127.0.0.1:8080"

# concurrency interval for lookups in miliseconds
interval = 200

# question cache capacity, 0 for infinite but not recommended (this is used for storing logs)
questioncachecap = 5000

# manual custom dns entries - comments for reference
customdnsrecords = [
    # "example.mywebsite.tld      IN A       10.0.0.1",
    # "example.other.tld          IN CNAME   wikipedia.org"
]

[Blocking]
    # response to blocked queries with a NXDOMAIN
    nxdomain = false
    # ipv4 address to forward blocked queries to
    nullroute = "0.0.0.0"
    # ipv6 address to forward blocked queries to
    nullroutev6 = "0:0:0:0:0:0:0:0"
    # manual blocklist entries
    blocklist = []
    # list of sources to pull blocklists from, stores them in ./sources
    sources = [
        "https://mirror1.malwaredomains.com/files/justdomains",
        "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
        "https://sysctl.org/cameleon/hosts",
        "https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
        "https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
        "https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-blocklist.txt"
    ]
    # list of locations to recursively read blocklists from (warning, every file found is assumed to be a hosts-file or domain list)
    sourcedirs = ["./sources"]
    sourcesStore = "./sources"
    # manual whitelist entries - comments for reference
    whitelist = [
        # "getsentry.com",
        # "www.getsentry.com"
    ]



[Upstream]
    # Dns over HTTPS provider to use.
    DoH = "https://cloudflare-dns.com/dns-query"
    # nameservers to forward queries to
    nameservers = ["1.1.1.1:53", "1.0.0.1:53"]
    # query timeout for dns lookups in seconds
    timeout_s = 5
    # cache entry lifespan in seconds
    expire = 600
    # cache capacity, 0 for infinite
    maxcount = 0

# Prometheus metrics
[Metrics]
    enabled = false
    path = "/metrics"
    # see https://cottand.github.io/leng/Prometheus-Metrics.html
    highCardinalityEnabled = false
    histogramsEnabled = false
    resetPeriodMinutes = 60

[DnsOverHttpServer]
    enabled = false
    bind = "0.0.0.0:80"
    timeoutMs = 5000

# TLS config is not required for DoH if you have some proxy (ie, caddy, nginx, traefik...) manage HTTPS for you
    [DnsOverHttpServer.TLS]
        enabled = false
        certPath = ""
        keyPath = ""
        # if empty, system CAs will be used
        caPath = ""
```

The most up-to-date version can be found on [config.go](https://github.com/Cottand/leng/blob/master/config.go)
