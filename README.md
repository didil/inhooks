[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)
![CI](https://github.com/didil/inhooks/actions/workflows/ci.yml/badge.svg)


# Inhooks
Inhooks aims to be a lightweight incoming webhooks gateway solution. Written in Go and runnable as a single binary or docker container. Only requires a redis database for storage/queueing.

*Note: This software is still early in its development cycle / not battle-tested. Test thoroughly before using in production systems.*


## Features
- Receive HTTP Webhooks and Enqueue to redis
- Fanout messages to multiple HTTP targets
- Delayed processing
- Retries on failure with configurable maximum number of attempts and delay
- ... more features coming

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
        delay: 30s # delay processing by 30 seconds
      - id: sink-2
        type: http
        url: https://example.com/othertarget
        retryAfter: 2m # on error, retry after 2 minutes
        maxAttemps: 10 # maximum number of attempts
```

With this config, inhooks will listen to http POST requests to `/api/v1/ingest/source-1-slug`.

When a message is received, it is saved to the redis database. Then inhooks tries to send it to each of the urls defined in the sinks section of the config.
In case of failures, retries are attempted based on the sink config params.

If the config is modifed, the server must be restarted to load the new config.

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

### Run Dev Server
```shell
make run-dev
```

### Run tests
```shell
make test
```

### Run linter
```shell
make lint
```


## Contributing
Feel free to open new issues or PRs ! You can also reach out to the maintainer at the email address: adil-inhooks@ledidil.com

## License
Copyright (c) 2023 Adil H.

Inhooks is an Open Source project licensed under the terms of the LGPLv3 license.
Please see [LICENSE](LICENSE) for the full license text.