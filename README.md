[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)

# Inhooks



Example inhooks.yml config
``` yaml
flows:
  - id: flow-1
    source:
      id: source-1
      type: http
    sinks:
      - id: sink-1
        type: http
        url: https://example.com/sink
```


## License

Copyright (c) 2023 Adil H.

Inhooks is an Open Source project licensed under the terms of the LGPLv3 license.
Please see [LICENSE](LICENSE) for the full license text.