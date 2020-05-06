[![Build Status](https://travis-ci.org/foadnh/healthcheck.svg?branch=master)](https://travis-ci.org/foadnh/healthcheck)
[![Test Coverage](https://api.codeclimate.com/v1/badges/544df3f8fcbde9d0c605/test_coverage)](https://codeclimate.com/github/foadnh/healthcheck/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/544df3f8fcbde9d0c605/maintainability)](https://codeclimate.com/github/foadnh/healthcheck/maintainability)
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
