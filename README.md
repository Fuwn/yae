# ⛩️ Yae

Yae is a powerful yet minimal dependency manager intended for use with Nix,
which functions similar to [niv](https://github.com/nmattia/niv/) and [`npins`](https://github.com/andir/npins/).

<details closed>
  <summary>Why should I use Yae instead of niv or <a href="https://github.com/andir/npins/"><code>npins</code></a>?</summary>

  1. No BS helper Nix expressions are needed by Yae. niv and `npins` spit out unnecessary and mostly complicated Nix expressions by default in the form of a file that you need to keep in sync with their mainline source. This requires additional upgrade commands from the CLI to maintain. A Yae environment is a single file and can be placed anywhere.
  2. Yae has a simple and coherent source tree. niv has a total of 10000 LOC, `npins` sits at almost 6000 LOC flat, and Yae stands at just shy of 1500 LOC when looking at all files. Yae's core source code itself sits at just 462 LOC, which is much, **much** smaller than that of niv and `npins`' core source trees. This is all to say that Yae implements everything needed to replace niv and `npins` in a much more efficient and concise codebase.
  3. Yae is simple by nature in design and usage philosophy.

     niv and `npins` are great, but are far too ~~overkill~~ overengineered for me and many other consumers. I say overengineered because I was able to write out Yae's initial implementation in just about thirty-minutes to an hour, and it was already complete enough for me to replace niv or `npins` in all of my production workflows. If you need some niche feature that niv or `npins` have, use them, but if not, Yae is here for you.
</details>

## Introduction

You can try out Yae without installing anything permanently on your system by running
`nix run github:Fuwn/yae`.

Check out [Tsutsumi](https://github.com/Fuwn/tsutsumi) to see an example of Yae running
in a production environment. Tsutsumi fully leverages the power of Yae to manage
and automagically update the sources of the Nix packages it provides using a simple
GitHub Actions CRON workflow.

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

# Adds a Yae dependency named `yaak` pinned at tag `2024.10.1` with tag trimming
# for updates
yae add \
  --type git \
  --unpack=false \
  --version 2024.10.1 \
  --trim-tag-prefix v \
  yaak \
  'https://github.com/yaakapp/app/releases/download/v{version}/yaak_{version}_amd64.AppImage.tar.gz'

# Updates all dependencies, e.g., updates the hash of `zen-browser-twilight-bin`
# and bumps the version of `zen-browser-bin` to `1.0.1-a.8`, handling URL and
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

### Integrating with Nix

To add Yae support to your Nix expression after running `yae init`, just read
from the Yae environment file. See the example below for more details.

### Nix Example

Here's an example snippet taken from Tsutsumi's [`zen-browser-bin` package](https://github.com/Fuwn/tsutsumi/blob/main/pkgs/zen-browser-bin.nix)
and [`yae.json`](https://github.com/Fuwn/tsutsumi/blob/main/yae.json#L59-L67)
showcasing Yae in action.

```nix
# pkgs/zen-browser-bin.nix

# This expression produces the `zen-browser-bin` package that Tsutsumi exposes
# as a Nix package derivation.
#
# Since it is managed by Yae, it is kept 100% up to date with zero effort through
# a Github Actions CRON job workflow that executes `yae update` periodically.
{
  pkgs,
  self,
  # This line imports Yae's environment configuration to be used below.
  yae ? builtins.fromJSON (builtins.readFile "${self}/yae.json"),
}:
# Tsutsumi exposes two versions of the Zen browser, the latest stable release
# and the latest Twilight release (a bleeding edge, daily build). This library
# function is one that takes one of two Yae sources for the Zen browser, and produces
# a Nix package derivation for it.
import "${self}/lib/zen-browser-bin.nix" {
  # Here, the latest SHA256 hash and release version from Yae are passed to Tsutsumi's
  # Zen browser packace function.
  #
  # If `yae update` is ran and a new release is detected, these values are
  # updated by Yae, which then triggers another workflow to build and send the
  # resulting derivation to Tsutsumi's binary cache.
  inherit (yae.zen-browser-bin) sha256 version;

  # To generate the Twilight release package, this is all that is changed.
  # inherit (yae.zen-browser-twilight-bin) sha256 version;
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
