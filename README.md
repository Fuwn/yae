# ⛩️ Yae

Yae is a simple dependency manager for use with Nix, similar to [niv](https://github.com/nmattia/niv/)
and [`npins`](https://github.com/andir/npins/).

I made it to solve my own problems, but I hope it can help you too.

## Introduction

You can try out Yae without installing anything permanently by running
`nix run github:Fuwn/yae`.

Additionally, [Tsutsumi](https://github.com/Fuwn/tsutsumi) uses Yae to manage
dependencies. You can check out a working implementation there.

## Usage

View the [installations instructions](#installation) below to set up Yae after
running `yae init`.

```sh
# Initialises a Yae environment in the current directory by creating an empty `yae.json`
# file
yae init

# Adds a Yae dependency named `zen-browser-twilight-bin` using a floating tag
# (tag always remain `twilight`, but may receive frequent hash changes)
yae add \
  --type binary \
  --version twilight \
  --unpack \
  zen-browser-twilight-bin \
  'https://github.com/zen-browser/desktop/releases/download/{version}/zen.linux-specific.tar.bz2'

# Adds a Yae dependency named `zen-browser-bin` pinned at tag `1.0.1-a.7`
yae add \
  --type git \
  --version 1.0.1-a.7 \
  --unpack \
  zen-browser-bin \
  'https://github.com/zen-browser/desktop/releases/download/{version}/zen.linux-specific.tar.bz2'

# Adds a Yae dependency named `yaak` pinned at tag `2024.10.1` with tag trimming for updates
yae add \
  --type git \
  --unpack=false \
  --version 2024.10.1 \
  --trim-tag-prefix v \
  yaak \
  'https://github.com/yaakapp/app/releases/download/v{version}/yaak_{version}_amd64.AppImage.tar.gz'

# Updates all dependencies, e.g., updates the hash of `zen-browser-twilight-bin`
# and bumps the version of `zen-browser-bin` to `1.0.1-a.8`, handling URI and
# hash recalculations, etc.
yae update

# Only updates `zen-browser-twilight-bin`
yae update zen-browser-twilight-bin
```

## Installation

Follow the installation instructions at [Tsutsumi](https://github.com/Fuwn/tsutsumi),
which provides both flake and flake-less installation options.

Alternatively, without flake-less support, install the
`inputs.yae.packages.${pkgs.system}.yae` package exposed by this flake.

### Nix

To add Yae support to your Nix expression after running `yae init`, just read
from the Yae environment file.

Here's an example taken from [Tsutsumi's `zen-browser-bin` package](https://github.com/Fuwn/tsutsumi/blob/main/pkgs/zen-browser-bin.nix).

```nix
# pkgs/zen-browser-bin.nix

{
  pkgs,
  self,
  yae ? builtins.fromJSON (builtins.readFile "${self}/yae.json"),
}:
import "${self}/lib/zen-browser-bin.nix" {
  # Effortless dependency updates just by running `yae update` from CI using a
  # CRON job
  inherit (yae.zen-browser-bin) sha256 version;
} { inherit pkgs; }
```

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
