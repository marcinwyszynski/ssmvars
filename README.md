# ssmvars

Utility library to store app configuration secrets using AWS SSM [Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) backend.

The concept behind this library is that the `VariableRepository` takes "exclusive ownership" of a variable `prefix`, and then allows setting up multiple `namespaces` as well as individual variables in them.

Resulting variables thus constructed like this in Parameter Store:

```text
/${prefix}/variables/${namespace}/${variable}
```

There are two types of variables - plain (read-write) and secret (write-only).
