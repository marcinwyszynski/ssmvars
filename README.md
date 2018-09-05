# ssmvars

[![Godoc](https://godoc.org/github.com/marcinwyszynski/ssmvars?status.svg)](http://godoc.org/github.com/marcinwyszynski/ssmvars)
[![CircleCI](https://circleci.com/gh/marcinwyszynski/ssmvars/tree/master.svg?style=svg)](https://circleci.com/gh/marcinwyszynski/ssmvars/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/marcinwyszynski/ssmvars)](https://goreportcard.com/report/github.com/marcinwyszynski/ssmvars)
[![codecov](https://codecov.io/gh/marcinwyszynski/ssmvars/branch/master/graph/badge.svg)](https://codecov.io/gh/marcinwyszynski/ssmvars)

Utility library to store app configuration secrets using AWS SSM [Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) backend.

The concept behind this library is that the `VariableRepository` takes "exclusive ownership" of a variable `prefix`, and then allows setting up multiple `namespaces` as well as individual variables in them.

Resulting variables thus constructed like this in Parameter Store:

```text
/${prefix}/variables/${namespace}/${variable}
```

There are two types of variables - plain (read-write) and secret (write-only). The latter type is implemented using [`SecureString`](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-about.html) - to manage those, you will need to provide a valid KMS key ID for the library to use.
