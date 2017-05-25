# masaba

masaba(真鯖) is a tool to show Mackerel system metrics in your terminal. The tool wraps [mkr](https://github.com/mackerelio/mkr) command because [mackerel-client-go](https://github.com/mackerelio/mackerel-client-go) is under development.

## Installation

git clone && make && make install. The binary will be installed to /usr/local/bin/masaba.

## How to use

### Run

```
export MACKEREL_APIKEY=<mackerel api key>
masaba -s SERVICE_NAME -r ROLE_NAME -i 1
```

### Options

* -s The service name(required).
* -r The role name(required).
* -i :The interval in second. (default 5)

## Image

![](image.png?raw=true)

## License

MIT
