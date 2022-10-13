# Cascade

Casade is a tool cascades soft deletes or hard deletes in your mongodb database. It best run as a daemon

## Installation, Getting Started

There are several ways to get started using Casade.

### Option 1: Download and run the docker image

```shell
$ docker run \
    -p 5005:5005 \
    --network=host \
    -v `pwd`/config.json:/config.json \
    ghcr.io/jirevwe/cascade:latest
```

### Option 2: Download a binary from the releases page

You can download the binary from the [releases](https://github.com/jirevwe/cascade/releases) page

### Option 3: Building cascde from source

To build Cascade from source code, you need:

- Go [v1.16 or greater](https://golang.org/doc/install)

Build it yourself

```shell
$ go build -o cascade ./cmd/*.go
```

Or install it

```
$ go install github.com/jirevwe/cascade@latest
```

#### Verify the install

```shell
$ cascade version
// v0.1.0
```

## Usage

Create a config file, this is a sample config file

```json
{
  "mongo_dsn": "mongodb://localhost:27017/convoy?rs=localhost",
  "db_name": "test",
  "redis_dsn": "redis://localhost:6379/1",
  "port": 4400,
  "relations": [
    {
      "parent": {
        "name": "users",
        "pk": "uid"
      },
      "children": [
        {
          "name": "wallets",
          "fk": "user_id"
        },
        {
          "name": "transactions",
          "fk": "sender_id"
        }
      ],
      "on": "replace",
      "do": "soft_delete"
    }
  ]
}
```

## Versioning

The CLI is versioned with [SemVer v2.0.0](https://semver.org/spec/v2.0.0.html). Releases are tagged with `vMAJOR.MINOR.PATCH` and published on the [Github releases page](https://github.com/jirevwe/cascade/releases).

## Usage manual

```shell
$ cascade
cascade your updates and deletes in any mongodb store

Usage:
  cascade [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  server      Starts the http server
  version     Print out the cli version

Flags:
  -h, --help   help for cascade

Use "cascade [command] --help" for more information about a command.
```
