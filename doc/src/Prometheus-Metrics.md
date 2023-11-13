The HTTP API has a `/metrics` endpoint that exposes Go runtime metrics as well as things including:
- downstream DNS requests, broken down by type
- upstream DNS requests
- upstream DNS-over-HTTPS success rate
- downstream DNS-over-HTTPS success rate


No grafana dashboards exist for leng yet. If you make one, please make a PR!