# grpctl

A [grpc_cli](https://grpc.github.io/grpc/core/md_doc_command_line_tool.html) inspired utility that attempts to provide a more natural command-line experience for any [gRPC](https://grpc.io/) service.

## Usage

### Creating a service

Before it can talk to a gRPC service, `grpctl` needs to know where to find it.

Create a service and set the address used:

```sh
$ grpctl service set snoot address localhost:50051
```

Here we've configured the `snoot` service to connect to `localhost:50051`.

> Note: the name "service" is reserved and cannot be used as the name of a user set service.

### Interacting with a service

Once an address is registered for a service, its name can be used as a direct subcommand of `grpctl`.

Following along with the previous example, run the new `snoot` service subcommand with the `--help` option to see what procedures it provides:

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

By default, `grpctl` will try to use [gRPC server reflection](https://grpc.github.io/grpc/core/md_doc_server-reflection.html) to discover available procedures at runtime, but it can also be configured to use a local `.proto` file as well.

Source the `snoot` service's service definitions from `snoot.proto`:

```sh
$ grpctl service set snoot proto snoot.proto
```

