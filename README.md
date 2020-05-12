[![Build/Test Status](https://circleci.com/gh/foadnh/healthcheck/tree/master.svg?style=svg)](https://circleci.com/gh/foadnh/healthcheck/tree/master)
[![Test Coverage](https://codecov.io/gh/foadnh/healthcheck/branch/master/graph/badge.svg)](https://codecov.io/gh/foadnh/healthcheck)
[![Go Report Card](https://goreportcard.com/badge/github.com/foadnh/healthcheck)](https://goreportcard.com/report/github.com/foadnh/healthcheck)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ffoadnh%2Fhealthcheck.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ffoadnh%2Fhealthcheck?ref=badge_shield)
# healthcheck 
Healthcheck for Go services

## Features
- Support background checks.
  - To protect services with fewer checks.
  - To improve response time of health check request.
- Support threshold for number of errors in the row.

## Motivation
Other implementations, has one of these 2 issues:
- Don't support background checks.
- Run a go routine per background check.

## Examples
`checkers` package has some examples.
