# grpctl

A [grpc_cli](https://grpc.github.io/grpc/core/md_doc_command_line_tool.html) inspired utility that attempts to provide a more natural command-line experience for any [gRPC](https://grpc.io/) service.

## Usage

### Creating a context

Before it can talk to a gRPC service, `grpctl` needs to know where to find it.

Create a context and set the address used:

```sh
$ grpctl context snoot set address localhost:50051
```

Here we've configured the `snoot` context to connect to `localhost:50051`.

> Note: the name "context" is reserved and cannot be used as the name of a user set context.

### Interacting with a service

Once a context has been created, it can be used as a direct subcommand of `grpctl`.

Run the service subcommand with the `--help` option to see what procedures it provides:

```sh
$ grpctl snoot --help
my favorite service!

Usage:
  grpctl snoot [command]

Available Commands:
  boop        boop a snoot
  list        list all snoots

Flags:
  -h, --help  help for snoot
```

### Command discovery

By default, `grpctl` will try to use [gRPC server reflection](https://grpc.github.io/grpc/core/md_doc_server-reflection.html) to discover available procedures at runtime, but it can also be configured to use a local protoc file as well.

```sh
$ grpctl context my-service set protoc my-service.protoc
```

