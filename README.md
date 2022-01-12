# Puma Exporter

Prometheus exporter for [Puma](https://github.com/puma/puma).

- [Forked from original repo](https://github.com:sapcc/puma-exporter)

## Requirements

- go 1.17+
- `- docker (buildx for multiarch docker images)`
- ``
- `## Build` default value

```
make build
```

## Environement variables

- `CONTROL_URL` default value `http://127.0.0.1:9293` - [control server](https://github.com/puma/puma#controlstatus-server)
- `AUTH_TOKEN` no default value
- `BIND_ADDRESS` default value `0.0.0.0:9882` - [prometheus default ports allocation](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)
