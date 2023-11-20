Here are some useful guides and resources for working with leng. Contributions welcome!


# Why Leng

Reasons you would want to use Leng include:
- **Ad-blocking at the DNS level**: this compliments misses browser adblockers
  (they use a different approach to block ads), and is especially useful in devices where
  ad-blockers are hard to install (like smart TVs, or non-browser apps).
- **Blocking tracking at the DNS level**: vendors, especially your
device's manufacturers, will often track you outside of websites (where browser ad-blockers
are powerless). When using the right blocklists, leng will block this
tracking for all devices that use it as their DNS provider.
- **DNS Server for self-hosted infra**: by specifying
your records on a config file, leng is a very easily maintanable
custom DNS server deployment.
- **DNS Privacy and Security**: many devices use the most basic DNS implementation, DNS over UDP.
This is a bad idea because it is less private and less secure ([you can read here
to understand why](https://www.cloudflare.com/en-gb/learning/dns/dns-over-tls/)). Leng can serve as a secure
proxy so that even if your devices speak to it via UDP, it speaks to the rest
of the internet via the more secure alternatives (like DoH or DoT).
- **It's small and fast**
- **There are few open-source DNS servers with the above features**:
my motivation for forking _grimd_ and creating leng was the need for a server
that provided blocklists (like _Blocky_) as well as decent custom DNS records
support (like _CoreDNS_, _grimd_ was almost there).

For more on leveraging leng for DNS privacy, see [DNS Privacy](DNS/Privacy.md).
