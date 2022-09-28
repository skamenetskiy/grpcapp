# grpcapp [![test](https://github.com/skamenetskiy/grpcapp/actions/workflows/go.yml/badge.svg)](https://github.com/skamenetskiy/grpcapp/actions/workflows/go.yml) [![Coverage Status](https://coveralls.io/repos/github/skamenetskiy/grpcapp/badge.svg?branch=main)](https://coveralls.io/github/skamenetskiy/grpcapp?branch=main) [![CodeQL](https://github.com/skamenetskiy/grpcapp/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/skamenetskiy/grpcapp/actions/workflows/codeql-analysis.yml) [![report](https://goreportcard.com/badge/github.com/skamenetskiy/grpcapp)](https://goreportcard.com/report/github.com/skamenetskiy/grpcapp) [![godoc](https://godoc.org/github.com/skamenetskiy/grpcapp?status.svg)](http://godoc.org/github.com/skamenetskiy/grpcapp)

A simple wrapper for gRPC microservices.

## Features

- Built-in CLI generator.
- Tiny and elegant.
- Includes exactly what's required.

## CLI Tool

To install the `grpcapp` cli tool run:
```
go install github.com/skamenetskiy/grpcapp/grpcapp@v0.0.3
```

For more help run: 
```
grpcapp help
```
```
usage: grpcapp {command} [...options]

commands:
	create {name} - create new application
	generate      - generate proto
	help          - print help information
```

## License

The grpcapp package is open-sourced software licensed under the [MIT license](LICENSE).