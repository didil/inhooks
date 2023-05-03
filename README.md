[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)

# Inhooks


## Inhooks config

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
## License

Copyright (c) 2023 Adil H.

Inhooks is an Open Source project licensed under the terms of the LGPLv3 license.
Please see [LICENSE](LICENSE) for the full license text.