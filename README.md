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
Inhooks consists of two major concepts, Sources and Sinks. A Source is an endpoint for receiving webhooks, and a [Sink](https://en.wikipedia.org/wiki/Sink_(computing)) is a target that receives the webhooks.

Inhooks listens to HTTP webhooks and saves the messages to Redis. A processing module retrieves the messages and sends them reliably to the defined sinks.

## Features
- Receive HTTP Webhooks and save them to a Redis database
- Fanout messages to multiple HTTP targets (sinks)
- Fast, concurrent processing
- Supports delayed processing
- Supports retries on failure with configurable number of attempts, interval and constant or exponential backoff
- Supports different HTTP payloads types: JSON, x-www-form-urlencoded, multipart/form-data
- Supports message transformation using JavaScript ECMAScript 5.1
- ... more features planned

## Downloading release binaries
The release binaries are available on the [GitHub releases](https://github.com/didil/inhooks/releases) page.
To download a specific version, adjust then env vars below and run:
```shell
export INHOOKS_VERSION="0.1.11"
export OS="linux"
export ARCH="amd64"
curl -LO https://github.com/didil/inhooks/releases/download/v${INHOOKS_VERSION}/inhooks_${INHOOKS_VERSION}_${OS}_${ARCH}.tar.gz
tar -xvzf inhooks_${INHOOKS_VERSION}_${OS}_${ARCH}.tar.gz
```

## Docker images:
The released docker images are available at: https://github.com/didil/inhooks/pkgs/container/inhooks

A minimal example to run the docker image:
```
docker run -it --rm \
-e REDIS_INHOOKS_DB_NAME=myinhooksdb  \
-e REDIS_URL=redis://<redis_user>:<redis_password>@<redis_host>:<redis_port> \
--mount type=bind,source="$(pwd)"/inhooks.yml,target=/app/inhooks.yml,readonly \
-p 3000:3000 \
ghcr.io/didil/inhooks:v0.1.11
```

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

### Env vars
Copy the .env examples to init the .env file and update as needed (to set the inhooks config file path, the redis url, the server port, etc).
```shell
cp .env.example .env
```

#### All env vars

| Category | Env var | Default | Description |
|---|---|---|---|
| **App** | `APP_ENV` | | Application environment (development, test, production) |
| | `INHOOKS_CONFIG_FILE` | `inhooks.yml` | Path to the flows config file |
| **Server** | `HOST` | `localhost` | Server bind host |
| | `PORT` | `3000` | Server bind port |
| | `SERVER_SHUTDOWN_GRACE_PERIOD` | `5s` | Graceful shutdown timeout |
| **Redis** | `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| | `REDIS_INHOOKS_DB_NAME` | | **Required.** Redis key prefix namespace |
| **Supervisor** | `SUPERVISOR_READY_WAIT_TIME` | `5s` | Delay before processing a newly enqueued message |
| | `SUPERVISOR_READY_QUEUE_CONCURRENCY` | `5` | Number of concurrent message processing goroutines |
| | `SUPERVISOR_ERR_SLEEP_TIME` | `5s` | Sleep time on processing error |
| | `SUPERVISOR_SCHEDULER_INTERVAL` | `30s` | Interval between scheduler runs |
| | `SUPERVISOR_PROCESSING_RECOVERY_INTERVAL` | `5m` | Interval to recover stuck processing messages |
| | `SUPERVISOR_QUEUE_METRICS_INTERVAL` | `30s` | Interval between queue metrics collection |
| | `SUPERVISOR_DONE_QUEUE_CLEANUP_ENABLED` | `false` | Enable automatic done queue cleanup |
| | `SUPERVISOR_DONE_QUEUE_CLEANUP_DELAY` | `336h` (14 days) | Age threshold for done message deletion |
| | `SUPERVISOR_DONE_QUEUE_CLEANUP_INTERVAL` | `60m` | Interval between done queue cleanup runs |
| | `SUPERVISOR_DEAD_QUEUE_CLEANUP_ENABLED` | `false` | Enable automatic dead queue cleanup |
| | `SUPERVISOR_DEAD_QUEUE_CLEANUP_DELAY` | `336h` (14 days) | Age threshold for dead message deletion |
| | `SUPERVISOR_DEAD_QUEUE_CLEANUP_INTERVAL` | `70m` | Interval between dead queue cleanup runs |
| **HTTP Client** | `HTTP_CLIENT_TIMEOUT` | `5s` | Timeout for outbound HTTP requests to sinks |
| **Sink** | `SINK_DEFAULT_DELAY` | `0` | Default delay before processing a message |
| | `SINK_DEFAULT_MAX_ATTEMPTS` | `3` | Default max delivery attempts per message |
| | `SINK_DEFAULT_RETRY_AFTER` | `0` | Default retry interval |
| | `SINK_DEFAULT_RETRY_EXP_MULTIPLIER` | `1` | Default retry exponential backoff multiplier |
| **Transform** | `TRANSFORM_JAVASCRIPT_TIMEOUT` | `1s` | Timeout for JavaScript transform execution |

### Queue cleanup
Inhooks can automatically delete old messages from the **done** and **dead** queues. Both cleanup features are **disabled by default** to prevent accidental data loss.

**Done queue cleanup** removes messages that have been successfully delivered and are older than the configured delay.

**Dead queue cleanup** removes messages that have exhausted all delivery attempts and are older than the configured delay. The age is determined from the timestamp of the message's last delivery attempt.

To enable, set the corresponding env vars:
```shell
SUPERVISOR_DONE_QUEUE_CLEANUP_ENABLED=true
SUPERVISOR_DONE_QUEUE_CLEANUP_DELAY=336h  # 14 days (default)

SUPERVISOR_DEAD_QUEUE_CLEANUP_ENABLED=true
SUPERVISOR_DEAD_QUEUE_CLEANUP_DELAY=336h  # 14 days (default)
```

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

### Message transformation

#### Transform definition

Message transformation allows you to modify the payload and headers of messages before they are sent to the sinks (destinations). This can be useful for tasks such as adding or removing fields, changing the format of the data, or adding custom headers.

Currently, only JavaScript transformations are supported. The JavaScript function should be named `transform` and should take two parameters: `bodyStr` (the message body as a string) and `headers` (the message headers as a JSON object). The function should return an array with two elements: the transformed payload as a string and the transformed headers as a JSON object.
The `headers` fields has the following format:
```
{
  "header-name": ["value1", "value2"]
}
```

Only JavaScript ECMAScript 5.1 is supported at the moment. We use the [goja](https://github.com/dop251/goja) library to execute the JavaScript code. You can read about the limitations on goja's documentation pages.

Here is an example configuration:
```yaml
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
        transform:
          id: js-transform-1
transform_definitions:
  - id: js-transform-1
    type: javascript
    script: |
      function transform(bodyStr, headers) {
        const body = JSON.parse(bodyStr);

        // add a header
        headers["X-INHOOKS-TRANSFORMED"] = ["1"];
        // capitalize the message if present
        if (body.msg) {
          body.msg = body.msg.toUpperCase();
        }
        // delete a key from the body
        delete body.my_dummy_key;

        return [JSON.stringify(body), headers];
      }
#  alternatively, you can provide the script path
#   script_path: /path/to/script.js
```


#### Testing transform scripts

You can use the `/api/v1/transform` endpoint to test your transform scripts before adding them to your flow configuration. This endpoint allows you to simulate the transformation process and see the results immediately.

To use this endpoint, send a POST request with a JSON payload containing the following fields:
- `body`: The message body as a string
- `headers`: The message headers as a JSON object
- `transformDefinition`: An object containing the `type` and `script` of your transformation

Here's an example of how to use the `/api/v1/transform` endpoint:
```shell
curl -X POST http://localhost:3000/api/v1/transform \
  -H "Content-Type: application/json" \
  -d '{
        "body": "{\"msg\": \"hello world\", \"my_dummy_key\": \"value\"}",
        "headers": {"Content-Type": ["application/json"]},
        "transformDefinition": {
          "type": "javascript",
          "script": "function transform(bodyStr, headers) { const body = JSON.parse(bodyStr); headers[\"X-INHOOKS-TRANSFORMED\"] = [\"1\"]; if (body.msg) { body.msg = body.msg.toUpperCase(); } delete body.my_dummy_key; return [JSON.stringify(body), headers]; }"
        }
      }'
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
