# Alternatives Comparison

Leng overlaps with a few other solutions in terms of providing DNS sinkholing for advertisements, as well as custom DNS.
This pages aims to clarify how leng compares to these other solutions.

> TLDR: Leng is suitable for a simple DNS server that serves custom records and blocks ads.
> It is designed to be small and easily scriptable (like Blocky), whereas Adguard, PiHole, etc are more comprehensive
solutions that include many more features but are not stateless, and are likely to have a larger fingerprint.

This is by all means not a comprehensive list.
Note I have not tried every single of these alternatives, so some information might be outdated
or plain wrong - if so please submit a PR to correct it if you find it so.

| Trait                                  | Leng                        | Blocky                      | Adguard                                                                                | PiHole                                                           | CoreDNS                                             |
|----------------------------------------|-----------------------------|-----------------------------|----------------------------------------------------------------------------------------|------------------------------------------------------------------|-----------------------------------------------------|
| Blocklist-basd blocking (remote fetch) | ✅                           | ✅                           | ✅                                                                                      | ❌ ish                                                            | ❌                                                   |
| Custom DNS records support             | ✅                           | ❌ ish (only rewrites)       | [❌ ish](https://github.com/AdguardTeam/AdGuardHome/wiki/Configuration)                 | ✅ ish via dnsmasq (no templating)                                | ✅                                                   |
| RAM footprint                          | 50MB with traffic + DoH     |                             | [150MB](https://adguard.com/kb/adguard-for-windows/installation/#system--requirements) | [512MB](https://docs.pi-hole.net/main/prerequisites/#hardware)   | 250MB but depends heavily on plugins                |
| Ease of use                            | Config file                 | Config file                 | Config file + Web UI                                                                   | Web UI                                                           | Config file                                         |
| Parental controls                      | Through parental blocklists | Through parental blocklists | ✅                                                                                      | ✅                                                                | ❌                                                   |
| DNS-over-HTTPS server                  | ✅                           | ✅                           | ✅                                                                                      | ✅                                                                | ❌                                                   | 
| DNS-over-HTTPS upstream proxy          | ✅                           | ✅                           | ✅                                                                                      | ✅                                                                | ✅                                                   |
| Stateless (all config as files)        | ✅                           | ✅                           | ❌                                                                                      | ❌                                                                | ✅                                                   |
| Running rootless                       | ✅                           | ✅                           | ✅                                                                                      | ❌                                                                | ✅                                                   |
| Prometheus metrics API                 | ✅                           | ✅                           | ❌ [see PR](https://github.com/AdguardTeam/AdGuardHome/pull/2312)                       | ❌ but [exporter exists](https://github.com/eko/pihole-exporter/) | [✅ via plugin](https://coredns.io/plugins/metrics/) |
| Per device config                      | ❌                           | ✅ via client groups         | ✅                                                                                      | ✅                                                                | ✅ via plugins                                       |
| DHCP Server (Assigns IPs to devices)   | ❌                           | ❌                           | ✅                                                                                      | ✅                                                                | ❌                                                   |
| Fancy Web UI                           | ❌                           | ❌                           | ✅                                                                                      | ✅                                                                | ❌                                                   |






