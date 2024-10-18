# Examples

## [Nixpkgs](https://github.com/Fuwn/yae/tree/main/examples/nixpkgs)

This example showcases adding the Nixpkgs unstable branch as a Yae source,
consuming it within an example flake, and mirroring the `hello` package from
Nixpkgs as a flake output.

Note that the flake has no inputs. This is because Yae directly manages the
Nixpkgs source.

This example is extremely useful and is intended be adapted to suite the specific
needs of flake-less Nix configurations, like classic Nix shells and flake-less system
configurations.

## [Nixpkgs Simple](https://github.com/Fuwn/yae/tree/main/examples/nixpkgs-simple)

This example is functionally identical to the Nixpkgs example, with the exception
that it utilises `builtins.currentSystem` to populate the `nixpkgs.system`
attribute, requiring the `--impure` command-line flag.

This example is purely for the sake of example, since in a real-world scenario,
you'd likely use something similar to [flake-utils](https://github.com/numtide/flake-utils)
for multi-system output management and populating the `nixpkgs.system` attribute.
