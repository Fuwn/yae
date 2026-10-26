# ⛩️ Yae

Yae is a powerful yet minimal dependency manager intended for use with Nix,
which functions similar to [niv](https://github.com/nmattia/niv/) and [`npins`](https://github.com/andir/npins/).

<details closed>
  <summary>How does Yae differ from niv and <code>npins</code>?</summary>

  1. Yae needs no helper Nix expressions.

     niv and `npins` generate Nix expressions that you keep in sync with their
     mainline sources. A Yae environment is a single JSON file that can be placed
     anywhere and read with `builtins.fromJSON`.
  2. Yae has a small, coherent source tree focused on fetching sources and
     updating their versions and hashes.
  3. If you need a niche feature that niv or `npins` has, use them; if not, Yae
     is here for you.

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

### Example Environment Setup

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

# This command adds a Yae dependency named `zen-browser-bin` starting at tag `1.0.1-a.7`.
yae add \
  --type git \
  --version 1.0.1-a.7 \
  --unpack \
  zen-browser-bin \
  'https://github.com/zen-browser/desktop/releases/download/{version}/zen.linux-specific.tar.bz2'

# This command adds a Yae dependency named `yaak` starting at tag `2024.10.1` with tag
# trimming for updates.
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

Use `--pin` when adding a source to prevent version and content updates.

`git` sources follow remote tags. To follow a branch, add a `binary` source whose
URL names the branch (for example, `--version main` with an
`archive/{version}.tar.gz` URL); every `yae update` then recomputes its hashes.

## Installation

You can either install Yae through the flake that this repository exposes or
through [Tsutsumi](https://github.com/Fuwn/tsutsumi).

[Tsutsumi](https://github.com/Fuwn/tsutsumi) provides both flake and flake-less installation
options, while this repository only provides installation support through flakes
using the exported `inputs.yae.packages.${system}.yae` package.

<details closed>
  <summary>Click here to see a minimal Nix flake that exposes a development shell with Yae bundled.</summary>

```nix
# Enter the development shell using `nix develop --impure` (impure is used here because `nixpkgs` internally
# assigns `builtins.currentSystem` to `nixpkgs.system` for the sake of simplicity in this example)
{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  inputs.tsutsumi = {
    url = "github:Fuwn/tsutsumi";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { nixpkgs, tsutsumi, ... }:
    let
      pkgs = import nixpkgs { };
      inherit (pkgs.stdenv.hostPlatform) system;
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = [ tsutsumi.packages.${system}.yae ];
      };
    };
}
```

</details>

### Integrating with Nix

To add Yae support to your Nix expression after running `yae init`, just read
from the Yae environment file. See the example below for more details.

### Examples

Check out the [`examples/`](./examples) directory for a couple of great examples
of Yae managing different sources including Nixpkgs!

#### Real-world Example

Tsutsumi reads its [`yae.json`](https://github.com/Fuwn/tsutsumi/blob/main/yae.json)
once in its [`flake.nix`](https://github.com/Fuwn/tsutsumi/blob/main/flake.nix)
and passes each source's fields straight to the standard fetchers.

<details closed>
  <summary>Expand this!</summary>

```nix
# flake.nix (simplified excerpt)
yae = builtins.fromJSON (builtins.readFile "${self}/yae.json");

# Every Yae source carries a URL and an SRI hash, which is exactly what
# `fetchzip` needs.
yaePackage =
  name:
  let
    archive = pkgs.fetchzip {
      inherit (yae.${name}) url hash;
    };
  in
  ((import "${archive}/flake.nix").outputs { inherit nixpkgs; self = archive; })
  .packages.${system}.default;
```

A scheduled GitHub Actions workflow runs
`yae --silent update --output-formatted-updated-list` and uses the printed list,
such as `maple, rui, & yae`, as its commit message.

</details>

## Licence

Licensed under either of [Apache License, Version 2.0](LICENSE-APACHE) or
[MIT license](LICENSE-MIT) at your option.

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in this project by you, as defined in the Apache-2.0 license, shall
be dual licensed as above, without any additional terms or conditions.
