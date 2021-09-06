# grpctl

A [grpc_cli](https://grpc.github.io/grpc/core/md_doc_command_line_tool.html) inspired utility that attempts to provide a more natural command-line experience for any [gRPC](https://grpc.io/) service.

## Usage

### Configuring a server

Before it can talk to a gRPC server, `grpctl` needs to know where to find it.

Create a server configuration and set the address used:

```sh
$ grpctl config set snoot address localhost:50051
```

Here we've configured the `snoot` server to connect to `localhost:50051`.

> Note: the name "server" is reserved and cannot be used as the name of a user set server.

### Interacting with a server

Once an address is registered for a server, its name can be used as a direct subcommand of `grpctl`.

Following along with the previous example, run the new `snoot` server subcommand with the `--help` option to see what methods it provides:

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

### Method commands

The methods of a service exposed by a configured server are made direct subcommands of that server's command if they are unique among all other services exposed by that server.

### Service commands

The services exposed by a server are always made direct subcommands of that server's command.

Service commands are hidden unless they share method names with other services from the same server.

### Command discovery

By default, `grpctl` will try to use [gRPC server reflection](https://grpc.github.io/grpc/core/md_doc_server-reflection.html) to discover available services at runtime, but it can also be configured to use a local `.proto` file as well.

Source the `snoot` service's service definitions from `snoot.proto`:

```sh
$ grpctl config set snoot proto snoot.proto
```

