# Examples

Run `yae update` inside an example directory to refresh its Nixpkgs hash.
The source uses `type: binary` because the archive URL follows a moving branch;
Yae rehashes that URL instead of discovering release tags. Both examples share
one `yae.json` through a symbolic link.

A moving URL may no longer match the saved hash when downloaded on a fresh
machine. Run the update before evaluating the example in that case.

## [Nixpkgs](https://github.com/Fuwn/yae/tree/main/examples/nixpkgs)

This example showcases adding the Nixpkgs unstable branch as a Yae source,
consuming it within an example flake, and mirroring the `hello` package from
Nixpkgs as a flake output.

Note that the flake has no inputs. This is because Yae directly manages the
Nixpkgs source.

This example is extremely useful and is intended to be adapted to suit the specific
needs of flake-less Nix configurations, like classic Nix shells and flake-less system
configurations.

## [Nixpkgs Simple](https://github.com/Fuwn/yae/tree/main/examples/nixpkgs-simple)

This example is functionally identical to the Nixpkgs example, with the exception
that it utilises `builtins.currentSystem` to populate the `nixpkgs.system`
attribute, requiring the `--impure` command-line flag.

This example is purely for the sake of example, since in a real-world scenario,
you'd likely use something similar to [flake-utils](https://github.com/numtide/flake-utils)
for multi-system output management and populating the `nixpkgs.system` attribute.
