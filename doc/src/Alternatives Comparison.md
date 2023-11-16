# Alternatives Comparison

Leng overlaps with a few other solutions in terms of providing DNS sinkholing for advertisements, as well as custom DNS.
This pages aims to clarify how leng compares to these other solutions.

> TLDR: Leng is suitable for a simple DNS server that serves custom records and blocks ads - not as a DHCP server, or for
monitoring
beyond Prometheus metrics, or for users that would like to perform admin through a web UI.

This is by all means not a comprehensive list.
Note I have not tried every single of these alternatives, so some information might be outdated
or plain wrong - if so please submit a PR to correct it if you find it so.

| Trait                                  | Leng                        | Adguard                                                                                | PiHole                                                         | CoreDNS                              | Blocky                      |
|----------------------------------------|-----------------------------|----------------------------------------------------------------------------------------|----------------------------------------------------------------|--------------------------------------|-----------------------------|
| Blocklist-basd blocking (remote fetch) | ✅                           | ✅                                                                                      | ❌ ish                                                          | ❌                                    | ✅                           |
| Custom DNS records support             | ✅                           | [❌ ish](https://github.com/AdguardTeam/AdGuardHome/wiki/Configuration)                 | ✅ ish via dnsmasq (no templating)                              | ✅                                    | ❌ ish (only rewrites)       |
| RAM footprint                          | 50MB with traffic + DoH     | [150MB](https://adguard.com/kb/adguard-for-windows/installation/#system--requirements) | [512MB](https://docs.pi-hole.net/main/prerequisites/#hardware) | 250MB but depends heavily on plugins |                             |
| Ease of use                            | Config file                 | Config file + Web UI                                                                   | Web UI                                                         | Config file                          | Config file                 |
| DHCP Server (Assigns IPs to devices)   | ❌                           | ✅                                                                                      | ✅                                                              | ❌                                    | ❌                           |
| Parental controls                      | Through parental blocklists | ✅                                                                                      | ✅                                                              | ❌                                    | Through parental blocklists |
| Per device config                      | ❌                           | ✅                                                                                      | ✅                                                              | ✅ via plugins                        | ✅ via client groups         |
| DNS-over-HTTPS server                  | ✅                           | ✅                                                                                      | ✅                                                              | ❌                                    | ✅                           | 
| DNS-over-HTTPS upstream proxy          | ✅                           | ✅                                                                                      | ✅                                                              | ✅                                    | ✅                           |
| Stateless (all config as files)        | ✅                           | ❌                                                                                      | ❌                                                              | ✅                                    | ✅                           |
| Fancy Web UI                           | ❌                           | ✅                                                                                      | ✅                                                              | ❌                                    | ❌                           |
| Running rootless                       | ✅                           | ✅                                                                                      | ❌                                                              | ✅                                    | ✅                           |



