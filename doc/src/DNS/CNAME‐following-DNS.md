# CNAME Following

Leng implements following CNAME records as specified in [RFC-1034 section 3.6.2](https://www.rfc-editor.org/rfc/rfc1034#section-3.6.2), where it returns all necessary CNAME and A records to fully resolve the query (as opposed to just returning a synthetic A record, which is known as [CNAME flattening](https://developers.cloudflare.com/dns/cname-flattening/cname-flattening-diagram/)).

> ⚠ This is the behaviour of most if not all DNS servers - Leng is only special in this in that it has to deal with cuustom DNS records, the resolvers it proxies, and blocklists. **You should not need to change its default behaviour**, but this page aims to leave it well-documented.

## `dig` request example

```bash
$> dig first.example

; <<>> DiG 9.18.19 <<>> first.example

;; QUESTION SECTION:
;first.example.		IN	A

;; ANSWER SECTION:
first.example.  	300	IN	CNAME	second.example.
second.example.		300	IN	CNAME	third.example.
third.example.  	300	IN	A	139.201.133.245
```

## Behaviour

The resolving for the downstream CNAME records is done with the same question type as the original question. That is, if you ask `AAAA some-cname.com`, the following CNAME queries will be `AAAA` questions too.

### Custom records

If you have set up your own custom records, those can also be part of the CNAME chain.

This makes it easy to alias custom records to external domains:
```
customdnsrecords = [
  "login.vpn       IN CNAME    this.very.long.other.domain.login.login.my-company.xyz"
]
```

Querying `login.vpn` will also return the A record corresponding to `this.very.long.other.domain.login.login.my-company.xyz`.

### Blocking
If any of the domains involved in the CNAME-following is part of a blocklist (that is, it would get blocked if it corresponded to an `A` response, rather than `CNAME`) then the entire request blocked (_unless_ the domain is part of the custom DNS defined in the config)

For the example where we have
```
first.example   IN CNAME second.example
second.example  IN CNAME third.example
third.example   IN A     10.0.0.0
```

if any of `first.example`, `second.example` or `third.example` appear in a blocklist, the request for `first.example` will fail.

## Configuration

CNAME-following is enabled by default, but you can disabled with the following:

```toml
# leng.toml

followCnameDepth = 0
```