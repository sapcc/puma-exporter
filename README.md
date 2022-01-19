# Puma Exporter

Prometheus exporter for [Puma](https://github.com/puma/puma).

- [Forked from original repo](https://github.com:sapcc/puma-exporter)

## Requirements

- go 1.17+
- docker - [buildx for multiarch docker images](https://docs.docker.com/buildx/working-with-buildx/#build-multi-platform-images)

## Environment variables

- `CONTROL_URL` default value `http://127.0.0.1:9293` - [control server](https://github.com/puma/puma#controlstatus-server)
- `AUTH_TOKEN` no default value
- `BIND_ADDRESS` default value `0.0.0.0:9882` - [prometheus default ports allocation](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)

## Build

```
make build
```

## Docker image

```
make docker
```

## Github Actions

- just checking that code is able to compile

## Local development

### Install ruby dependencies

```
bundle install
```

###  Run puma locally - will used `config/puma.rb`

```
bundle exec puma
```

### Build and run exporter (another terminal)

```
go build
./bin/puma_exporter
```

### Test it

Prometheus metrics by `puma_exporter`

```
curl http://127.0.0.1:9882/metrics
```

Prometheus metrics by `puma-metrics` plugin

```
curl http://127.0.0.1:9393/metrics
```

Results should be the same.

## TODO

- [ ] reserve port for exporter in official Prometheus documentation
- [ ] add support for multiarch docker images
