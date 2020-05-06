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
