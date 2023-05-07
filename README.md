[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)
![CI](https://github.com/didil/inhooks/actions/workflows/ci.yml/badge.svg)


# Inhooks
Inhooks aims to be a lightweight incoming webhooks gateway solution. Written in Go and runnable as a single binary or docker container. Requires a redis database for storage/queueing.

## Inhooks config
The inhooks config file allows setting up the Source to Sink flows

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
        url: https://example.com/sink
        delay: 30s # delay processing by 30 seconds
```

## Development setup
### Tools
Install Go 1.20+ and Redis 6.2.6+


Install tools
```shell
make install-tools
```

### Env
Copy the .env examples to init your .env files
```shell
cp .env.example .env
cp .env.test.example .env.test
```

## Contributing
Feel free to open new issues or PRs ! You can also reach out to the maintainer at the email address: adil-inhooks@ledidil.com

## License

Copyright (c) 2023 Adil H.

Inhooks is an Open Source project licensed under the terms of the LGPLv3 license.
Please see [LICENSE](LICENSE) for the full license text.