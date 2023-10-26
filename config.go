package main

import (
	"fmt"
	"github.com/jonboulle/clockwork"
	"github.com/pelletier/go-toml/v2"
	"log"
	"os"
)

// BuildVersion returns the build version of grimd, this should be incremented every new release
var BuildVersion = "2.2.1"

// ConfigVersion returns the version of grimd, this should be incremented every time the config changes so grimd presents a warning
var ConfigVersion = "2.2.1"

// Config holds the configuration parameters
type Config struct {
	Version           string
	Sources           []string
	SourceDirs        []string
	LogConfig         string
	Bind              string
	API               string
	NXDomain          bool
	Nullroute         string
	Nullroutev6       string
	Nameservers       []string
	Interval          int
	Timeout           int
	Expire            uint32
	Maxcount          int
	QuestionCacheCap  int
	TTL               uint32
	Blocklist         []string
	Whitelist         []string
	CustomDNSRecords  []string
	ToggleName        string
	ReactivationDelay uint
	APIDebug          bool
	DoH               string
	Metrics           Metrics `toml:"metrics"`
}

type Metrics struct {
	Enabled bool
	Path    string
}

var defaultConfig = `
# version this config was generated from
version = "%s"

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
# e.g. logconfig = "file:grimd.log@2,syslog@1,stderr@2"
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

# When this string is queried, toggle grimd on and off
togglename = ""

# If not zero, the delay in seconds before grimd automaticall reactivates after
# having been turned off.
reactivationdelay = 300

#Dns over HTTPS provider to use.
DoH = "https://cloudflare-dns.com/dns-query"

# Prometheus metrics - enable 
[Metrics]
  enabled = false
  path = "/metrics"
`

func parseDefaultConfig() Config {
	var config Config

	err := toml.Unmarshal([]byte(defaultConfig), &config)
	if err != nil {
		logger.Fatalf("There was an error parsing the default config %v", err)
	}
	config.Version = ConfigVersion

	return config
}

// WallClock is the wall clock
var WallClock = clockwork.NewRealClock()

// LoadConfig loads the given config file
func LoadConfig(path string) (*Config, error) {

	var config = parseDefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("warning, config not found - using defaults")
		return &config, nil
	}

	if err := toml.Unmarshal([]byte(path), &config); err != nil {
		return nil, fmt.Errorf("could not load config: %s", err)
	}

	if config.Version != ConfigVersion {
		if config.Version == "" {
			config.Version = "none"
		}

		log.Printf("warning, grimd.toml is out of date!\nconfig v%s\ngrimd config v%s\ngrimd v%s\nplease update your config\n", config.Version, ConfigVersion, BuildVersion)
	} else {
		log.Printf("grimd v%s\n", BuildVersion)
	}

	return &config, nil
}
