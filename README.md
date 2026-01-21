prowl
=====

This is a server that delivers system usage information.

## How to build

Build using the standard `go build` command.

## How to run

Running on a Linux eg. server, simply run the `./prowl` command.

## Configuring

*prowl* supports a few command line parameters:

* `-port [port_number]` lets you set the port number the server should listen on. Default is `5000`.
* `-protect` set it to require a query parameter with a _secret key_ (the key has to be set, either with the `-secret` flag or the `PROWL_SECRET` environment variable).
* `-secret [secret_key]` is to set the secret to use if you use the `-protect` flag.
* `-relay` enables relay mode so the server accepts POSTs at `/relay?machine_name=...`.
* `-relay-local` controls whether the relay server reports its own stats (default: true).
* `-relay-target [url]` posts stats to a relay server instead of serving locally.
* `-machine-name [name]` sets the machine name used when posting to a relay server.
* `-relay-secret [secret_key]` is the secret used when posting to a relay server (or set `PROWL_RELAY_SECRET`).
