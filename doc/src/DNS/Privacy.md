# DNS Privacy

Leng can enhance your DNS Privacy in several ways

## As your DoH provider

[DNS-over-HTTPS](https://www.cloudflare.com/en-gb/learning/dns/dns-over-tls/)
allows encrypted, hard-to-block DNS. You can set up DNS-over-HTTPS
for most major browsers ([see how here](https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-https/encrypted-dns-browsers/)).

See how to set it up for leng at [DNS-over-HTTP](DNS-over-HTTPS-(DoH).md).

If all you want is to use DoH, and you do not really care about ad/tracking-blocking,
using leng instead of an existing DoH provider directly has little
benefit: you will be getting
all the features of DoH, but the DNS provider will still know what you are visiting.
There isn't an easy way around this: we need to resolve your DNS query somehow!



## As a DoH proxy

DoH is great, but most devices use DNS-over-UDP by default, and some can't even
be configured otherwise.

If you have your own private secure network, you can stop
attackers from learning what websites you visit by using leng as
a secure proxy:

```mermaid

graph TD
    subgraph Secure Network
        U("🧘 User") --> |"🔓 Insecure\nDNS-over-UDP"|L[Leng]
    end
    L --> |"🔒 Secure DoH"| Up[Upstream DNS]
    A("👿 Attacker") ---> |Cannot see contents\nof DNS requests | Up
```

This way you allow 'insecure' DNS, but only inside your network,
and your requests are private to external attackers.

No configuration is required for this: leng will always try
to resolve domains by DoH via cloudflare before falling back to
other methods. You can choose the upstream DoH resolver in the
[Configuration](Configuration.md).

> Note that this method is only as secure as your network is!
> Ideally set up as many devices as possible to use DoH directly

