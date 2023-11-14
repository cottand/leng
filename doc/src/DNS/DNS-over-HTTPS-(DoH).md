# DNS-over-HTTP(S), aka DoH

Leng supports DNS-over-HTTPS as per [RFC-8484](https://datatracker.ietf.org/doc/html/rfc8484), although it is disabled by default.

Custom DNS records will be served over DoH the same as normal DNS requests.

You can specify your key files yourself to have leng serve HTTPS traffic, or you can let leng serve HTTP traffic and have a proxy manage the HTTPS certificates.

### Specifying Key files (HTTP)

```toml
[DnsOverHttpServer]
    enabled = true
    bind = "0.0.0.0:80"
    timeoutMs = 5000

    [DnsOverHttpServer.TLS]
        enabled = true
        certPath = ""
        keyPath = ""
        # if empty, system CAs will be used
        caPath = ""
```

### Not specifying key files (TLS disabled, HTTP traffic from leng)
```toml
[DnsOverHttpServer]
    enabled = true
    bind = "0.0.0.0:80"
    timeoutMs = 5000
```

> ⚠ It is not recommended to use HTTP without TLS at all directly. Your queries will be un-encrypted, so they won't be much different than normal UDP queries.

You can use DoH [in most browsers](https://ghacks.net/2021/10/23/how-to-enable-dns-over-https-secure-dns-in-chrome-brave-edge-firefox-and-other-browsers/).