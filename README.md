# ⛩️ Yae

Yae is a simple dependency manager for use with Nix, similar to [niv](https://github.com/nmattia/niv/)
and [`npins`](https://github.com/andir/npins/).

I made it to solve my own problems, but I hope it can help you too.

## Introduction

You can try out Yae without installing anything permanently by running
`nix run github:Fuwn/yae`.

Additionally, [Tsutsumi](https://github.com/Fuwn/tsutsumi) uses Yae to manage
dependencies. You can check out a working implementation there.

## Installation

Follow the installation instructions at [Tsutsumi](https://github.com/Fuwn/tsutsumi).

## `--help`

```text
NAME:
   yae - Nix Dependency Manager

USAGE:
   yae [global options] command [command options]

DESCRIPTION:
   Nix Dependency Manager

AUTHOR:
   Fuwn <contact@fuwn.me>

COMMANDS:
   init     Initialise a new Yae environment
   add      Add a source
   drop     Drop a source
   update   Update one or all sources
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --sources value  Sources path (default: "./yae.json")
   --help, -h       show help

COPYRIGHT:
   Copyright (c) 2024-2024 Fuwn
```

## Licence

This project is licensed with the [GNU General Public License v3.0](./LICENSE.txt).
