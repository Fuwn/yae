# 🦖 Wiene

Wiene is a simple dependency manager for use with Nix, similar to [niv](https://github.com/nmattia/niv/)
and [`npins`](https://github.com/andir/npins/).

I made it to solve my own problems, but I hope it can help you too.

Try it out without installing anything permanently by running
`nix run github:Fuwn/wiene`!

## Installation

Follow the installation instructions at [Tsutsumi](https://github.com/Fuwn/tsutsumi).

## `--help`

```text
NAME:
   wiene - Nix Dependency Manager

USAGE:
   wiene [global options] command [command options]

DESCRIPTION:
   Nix Dependency Manager

AUTHOR:
   Fuwn <contact@fuwn.me>

COMMANDS:
   update   Update one or all sources
   drop     Drop a source
   add      Add a source
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --sources value  Sources path (default: "./sources.json")
   --help, -h       show help

COPYRIGHT:
   Copyright (c) 2024-2024 Fuwn
```

## Licence

This project is licensed with the [GNU General Public License v3.0](./LICENSE.txt).
