# Puma Exporter

Puma Exporter is a Prometheus exporter. It provides metrics on Puma's performance, including request backlog and thread count, and serves these metrics via HTTP for Prometheus to scrape.

## Overview

This exporter is designed to query Puma's control server for performance metrics and expose them to Prometheus. Metrics provided include:

- `puma_request_backlog`: Number of requests waiting to be processed by a thread.
- `puma_thread_count`: Number of threads currently running.

## Getting Started

### Installation

Clone the repository and build the project:

```bash
git clone https://github.com/sapcc/puma-exporter.git
cd puma-exporter
make build
```

### Usage

Run the exporter:

```bash
./bin/puma-exporter
```

### Building and pushing a new Docker image

After making changes to the code, you can build and push a new Docker image:

1. Upgrade the version in the `Makefile`.
2. Run the following commands:

```bash
make docker
make push
```

## Upload the changes to the repo

```bash
git add -A
git commit -m "release new version 1.x.x"
git push
```

### Update the Helm Charts

https://github.com/sapcc/helm-charts/blob/master/openstack/elektra/templates/deployment.yaml#L118
