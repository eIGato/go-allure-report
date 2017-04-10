# WIP: go-allure-report

Converts `go test` output to an xml report, suitable for [Allure](http://allure.qatools.ru).

## Installation

Go version 1.1 or higher is required. Install or update using the `go get`
command:

```bash
go get -u github.com/eIGato/go-allure-report
```

## Usage

go-allure-report reads the `go test` verbose output from standard in and writes
allure compatible XML to standard out.

```bash
go test -v 2>&1 | go-allure-report > report.xml
```
