# Custom DNS Records

You can make leng return records of your choosing (which will take precedence over upstream DNS records) by setting `customdnsrecords` in the [Configuration](../Configuration.md).

Custom DNS records are represented as [Resource Record](https://en.wikipedia.org/wiki/Domain_Name_System#Resource_records) strings. Class defaults to IN and TTL defaults to 3600. Full zone file syntax is supported.

```toml
customdnsrecords = [
    "example.com.         3600 IN  A       10.10.0.1",
    "example.cname.com.        IN  CNAME   wikipedia.org",
]
```
