package main

import (
	cTls "crypto/tls"
	"errors"
	"fmt"
	"github.com/cottand/leng/tls"
	"github.com/jonboulle/clockwork"
	"github.com/pelletier/go-toml/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// BuildVersion returns the build version of leng, this should be incremented every new release
var BuildVersion = "1.4.0"

// ConfigVersion returns the version of leng, this should be incremented every time the config changes so leng presents a warning
var ConfigVersion = "1.4.0"

// Config holds the configuration parameters
type Config struct {
	Version           string
	LogConfig         string
	Bind              string
	API               string
	Interval          int
	Timeout           int
	QuestionCacheCap  int
	TTL               uint32
	CustomDNSRecords  []string
	APIDebug          bool
	Blocking          Blocking
	Upstream          Upstream
	Metrics           Metrics `toml:"metrics"`
	DnsOverHttpServer DnsOverHttpServer
	FollowCnameDepth  uint32
}

type Blocking struct {
	Sources      []string
	SourcesStore string
	SourceDirs   []string
	Blocklist    []string
	Whitelist    []string
	NXDomain     bool
	Nullroute    string
	Nullroutev6  string
}

type Upstream struct {
	DoH         string
	Nameservers []string
	TimeoutS    int `toml:"timeout_s"`
	Expire      uint32
	Maxcount    int
}

type Metrics struct {
	Enabled bool
	Path    string
}

type DnsOverHttpServer struct {
	Enabled   bool
	Bind      string
	TimeoutMs int64
	TLS       TlsConfig
	parsedTls *cTls.Config
}

type TlsConfig struct {
	certPath, keyPath, caPath string
	enabled                   bool
}

func (c TlsConfig) parsedConfig() (*cTls.Config, error) {
	if !c.enabled {
		return nil, nil
	}
	return tls.NewTLSConfig(c.certPath, c.keyPath, c.caPath)
}

var defaultConfig = `
# version this config was generated from
version = "%s"

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

# concurrency interval for lookups in milliseconds
interval = 200

# question cache capacity, 0 for infinite but not recommended (this is used for storing logs)
questioncachecap = 5000

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

# How deep to follow chains of CNAME records
# set to 0 to disable CNAME-following entirely
# (anything more than 10 should be more than plenty)
# see https://github.com/Cottand/leng/wiki/CNAME%E2%80%90following-DNS
followCnameDepth = 12

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

# Prometheus metrics - disabled by default
[Metrics]
	enabled = false
	path = "/metrics"

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
`

func parseDefaultConfig() Config {
	var config Config

	err := toml.Unmarshal([]byte(defaultConfig), &config)
	if err != nil {
		logger.Fatalf("There was an error parsing the default config: %v", err)
	}
	config.Version = ConfigVersion

	return config
}

// WallClock is the wall clock
var WallClock = clockwork.NewRealClock()

func contextualisedParsingErrorFrom(err error) error {
	errString := strings.Builder{}
	var derr *toml.DecodeError
	_, _ = fmt.Fprint(&errString, "could not load config:", err)
	if errors.As(err, &derr) {
		errString.WriteByte('\n')
		_, _ = fmt.Fprintln(&errString, derr.String())
		row, col := derr.Position()
		_, _ = fmt.Fprintln(&errString, "error occurred at row", row, "column", col)
	}
	return errors.New(errString.String())
}

// LoadConfig loads the given config file
func LoadConfig(path string) (*Config, error) {

	var config = parseDefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("warning, config not found - using defaults")
		return &config, nil
	}

	path = filepath.Clean(path)
	file, err := os.Open(path)
	if err != nil {
		log.Printf("warning, failed to open config (%v) - using defaults", err)
		return &config, nil
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	d := toml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, contextualisedParsingErrorFrom(err)
	}

	dohTls, err := config.DnsOverHttpServer.TLS.parsedConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load TLS config: %s", err)
	}
	config.DnsOverHttpServer.parsedTls = dohTls

	if config.Version != ConfigVersion {
		if config.Version == "" {
			config.Version = "none"
		}

		log.Printf("warning, leng.toml is out of date!\nconfig v%s\nleng config v%s\nleng v%s\nplease update your config\n", config.Version, ConfigVersion, BuildVersion)
	} else {
		log.Printf("leng v%s\n", BuildVersion)
	}

	return &config, nil
}
