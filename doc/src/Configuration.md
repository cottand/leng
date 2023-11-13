If leng.toml is not found the default configuration will be used. If it is found, fields that are set will act as overrides.

Here is the default configuration:

```toml
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
sourcedirs = [
	"sources"
]

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

# response to blocked queries with a NXDOMAIN
nxdomain = false

# ipv4 address to forward blocked queries to
nullroute = "0.0.0.0"

# ipv6 address to forward blocked queries to
nullroutev6 = "0:0:0:0:0:0:0:0"

# nameservers to forward queries to
nameservers = ["1.1.1.1:53", "1.0.0.1:53"]

# concurrency interval for lookups in miliseconds
interval = 200

# query timeout for dns lookups in seconds
timeout = 5

# cache entry lifespan in seconds
expire = 600

# cache capacity, 0 for infinite
maxcount = 0

# question cache capacity, 0 for infinite but not recommended (this is used for storing logs)
questioncachecap = 5000

# manual blocklist entries
blocklist = []

# Drbl related settings
usedrbl = 0
drblpeersfilename = "drblpeers.yaml"
drblblockweight = 128
drbltimeout = 30
drbldebug = 0

# manual whitelist entries - comments for reference
whitelist = [
	# "getsentry.com",
	# "www.getsentry.com"
]

# manual custom dns entries - comments for reference
customdnsrecords = [
    # "example.mywebsite.tld      IN A       10.0.0.1"
    # "example.other.tld          IN CNAME   wikipedia.org"
]

# When this string is queried, toggle leng on and off
togglename = ""

# If not zero, the delay in seconds before leng automaticall reactivates after
# having been turned off.
reactivationdelay = 300

#Dns over HTTPS provider to use.
DoH = "https://cloudflare-dns.com/dns-query"

# Prometheus metrics - enable 
[Metrics]
  enabled = false
  path = "/metrics"
```

The most up-to-date version can be found on [config.go](https://github.com/Cottand/leng/blob/master/config.go)