[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)
![CI](https://github.com/didil/inhooks/actions/workflows/ci.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/didil/inhooks/badge.svg?branch=main)](https://coveralls.io/github/didil/inhooks?branch=main)

# Inhooks

![Inhooks Logo](logo-no-background.png?raw=true "Inhooks Logo")

Inhooks aims to be a lightweight incoming webhooks gateway solution. Written in Go and runnable as a single binary or docker container. Only requires a redis database for storage/queueing.

You can read more about inhooks in the [launch blog post](https://didil.medium.com/inhooks-3983b68274e1).

*Note: This software is still early in its development cycle / not battle-tested. Test thoroughly before using in production systems.*

## Architecture

![Inhooks Architecture](inhooks-architecture.png?raw=true "Inhooks Architecture")

### High level overview
Inhooks listens to HTTP webhooks and saves the messages to Redis. A processing module retrieves the messages and sends them reliably to the defined targets.

## Features
- Receive HTTP Webhooks and save them to a Redis database
- Fanout messages to multiple HTTP targets
- Fast, concurrent processing
- Supports delayed processing
- Supports retries on failure with configurable number of attempts, interval and constant or exponential backoff
- ... more features planned

## Usage
### Inhooks config
The inhooks config file allows setting up the Source to Sink flows.
Create a file named inhooks.yml in the folder where the inhooks server will run (configurable via the INHOOKS_CONFIG_FILE env variable).

Example inhooks.yml config
``` yaml
flows:
  - id: flow-1
    source:
      id: source-1
      slug: source-1-slug
      type: http
    sinks:
      - id: sink-1
        type: http
        url: https://example.com/target
        delay: 90s # delay processing by 90 seconds
      - id: sink-2
        type: http
        url: https://example.com/othertarget
        retryInterval: 5m # on error, retry after 5 minutes
        # retryExpMultiplier: 2 # exponential backoff
        maxAttempts: 10 # maximum number of attempts
```

With this config, inhooks will listen to http POST requests to `/api/v1/ingest/source-1-slug`.

When a message is received, it is saved to the redis database. Then inhooks tries to send it to each of the urls defined in the sinks section of the config.
In case of failures, retries are attempted based on the sink config params.

If the config is modifed, the server must be restarted to load the new config.

### Securing webhooks
If you would like to verify your webhooks with HMAC 256, you can use the following configuration:

``` yaml
flows:
  - id: flow-1
    source:
      id: source-1
      slug: source-1-slug
      type: http
      verification:
        verificationType: hmac # only option supported at the moment
        hmacAlgorithm: sha256 # only option supported at the moment
        signatureHeader: x-my-header # the name of the http header in the incoming webhook that contains the signature
        signaturePrefix: "sha256=" # optional signature prefix that is required for some sources, such as github for example that uses the prefix 'sha256='
        currentSecretEnvVar: VERIFICATION_FLOW_1_CURRENT_SECRET  # the name of the environment variable containing the verification secret
        previousSecretEnvVar: VERIFICATION_FLOW_1_PREVIOUS_SECRET # optional env var that allows rotating secrets without service interruption
```
### Prometheus metrics
Inhooks exposes Prometheus metrics at the `/api/v1/metrics` endpoint.

## Development setup
### Tools
Go 1.20+ and Redis 6.2.6+ are required

Install tools
```shell
make install-tools
```

### Env
Copy the .env examples to init the .env files
```shell
cp .env.example .env
cp .env.test.example .env.test
```

### Run tests
```shell
make test
```

### Run linter
```shell
make lint
```

### Run Dev Server
```shell
make run-dev
```

### Run Docker Compose
```shell
docker-compose up
```

## Contributing
Feel free to open new issues or PRs ! You can also reach out to the maintainer at the email address: adil-inhooks@ledidil.com

## License
Copyright (c) 2023 Adil H.

Inhooks is an Open Source project licensed under the terms of the LGPLv3 license.
Please see [LICENSE](LICENSE) for the full license text.