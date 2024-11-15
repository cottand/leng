# Prometheus metrics

The HTTP API has a `/metrics` endpoint that exposes Go runtime metrics as well as things including:

- downstream DNS requests, broken down by type
- upstream DNS requests
- upstream DNS-over-HTTPS success rate
- downstream DNS-over-HTTPS success rate

No grafana dashboards exist for leng yet. If you make one, please make a PR!

## High cardinality metrics

Tags can be added to some metrics (`upstream_request`, `request_total`) so that
they include information such as the name of the DNS request (ie, `example.com.`)
or the IP of host making the request.

If leng is left to run for a few hours (and you have enough traffic),
the cardinality of these metrics will grow, to the point
the
size of the `/metrics`
response will grow to be so big the metrics stop being updated.
While resetting the counters periodically can help
(and you can tweak that with the config `Metrics.resetPeriodMinutes`)
you might still see issues depending on your traffic.
You can
read [this SO post](https://stackoverflow.com/questions/46373442/how-dangerous-are-high-cardinality-labels-in-prometheus)
to learn more.

High cardinality metrics **can also compromise your privacy** by exposing in the metrics endpoint
what domains clients are querying as well as their IPs.

For these reasons, **high cardinality metrics are disabled by default**. You can enable them
with the following config:

```toml
[Metrics]
enabled = true
path = "/metrics"
highCardinalityEnabled = true
```

## Histogram metrics

Histogram metrics are not unbounded and usually will not be as high-cardinality as the metrics discussed above,
but you should still expect them to have some impact on leng's the memory footprint.

You can enable them with:

```toml
[Metrics]
enabled = true
path = "/metrics"
histogramsEnabled = true
```
