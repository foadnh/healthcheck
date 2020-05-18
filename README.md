[![Build/Test Status](https://circleci.com/gh/foadnh/healthcheck/tree/master.svg?style=shield)](https://circleci.com/gh/foadnh/healthcheck/tree/master)
[![Test Coverage](https://codecov.io/gh/foadnh/healthcheck/branch/master/graph/badge.svg)](https://codecov.io/gh/foadnh/healthcheck)
[![Go Report Card](https://goreportcard.com/badge/github.com/foadnh/healthcheck)](https://goreportcard.com/report/github.com/foadnh/healthcheck)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ffoadnh%2Fhealthcheck.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ffoadnh%2Fhealthcheck?ref=badge_shield)

# healthcheck 
Healthcheck for Go services.

## Document
[pkg.go.dev](https://pkg.go.dev/github.com/foadnh/healthcheck)

## Features
- Support background checks.
  - To protect services with expensive checks.
  - To improve response time of health check request.
- Support threshold for number of errors in a row.
- A Detailed format.
  - By default, response do not have body.
  - Pass detail query parameter in the request for detailed response. Good for debugging.

## Motivation
Other implementations, has one of these 2 issues:
- Don't support background checks.
- Run a go routine per background check.

## Usage

### Running Service
- Create a new ServeMux. Or use `http.DefaultServeMux`.
```go
serveMux := http.NewServeMux()
```
- Create a new HealthCheck instance. Pass ServeMux and healthcheck path.
```go
h := healthcheck.New(serveMux, "/healthcheck")
```
- Register as many as _checks_ you have.
  - name: A unique name per _check_.
  - check: _Check_ function.
  - timeout: Timeout of check.
  - opts: Check options. Type `CheckOption`. [Checker Options section](#checker-options)
```go
h.Register("check 1", checkOne, time.Second)
h.Register("check 2", checkTwo, time.Second*10, InBackground(time.Minute*10))
```
- Run it (If you don't have background _checks_, no need for this step). Remember to close it.
```go
h.Run(context.Background())
defer h.Close()
```
### Creating Checkers
A _checker_ is a function with this signature:
```go
type Checker func(ctx context.Context) error
```
Healthcheck has built-in timeouts for _checks_, no need to add it in _checker_. `ctx` is with a timeout, use it to release resources if needed.

### Checker Options
Pass checker options when registering _checks_ to modify the behavior.

- **InBackground** forces a _check_ to run in the background.
```go
InBackground(interval time.Duration)
```
- **WithThreshold** adds a threshold of errors in the row to show unhealthy state.
```go
WithThreshold(threshold uint)
```

## Examples
For creating new Checks, [checkers package](checkers/README.md) has some examples.

Executable example in [pkg.go.dev](https://pkg.go.dev/github.com/foadnh/healthcheck).


## License
This project is licensed under the terms of the Mozilla Public License 2.0.
