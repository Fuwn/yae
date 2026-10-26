# Examples

Both examples read one `yae.json`, shared through a symbolic link, that tracks
the latest Nixpkgs release tag. Run `yae update` inside an example directory to
move it to a newer release; Yae finds the tag, rewrites the URL, and recomputes
both hashes.

## [Nixpkgs](https://github.com/Fuwn/yae/tree/main/examples/nixpkgs)

This example consumes the Yae-managed Nixpkgs source within a flake and mirrors
the `hello` package from Nixpkgs as a flake output.

Note that the flake has no inputs. This is because Yae directly manages the
Nixpkgs source.

The same pattern suits flake-less Nix configurations, like classic Nix shells
and flake-less system configurations.

## [Nixpkgs Simple](https://github.com/Fuwn/yae/tree/main/examples/nixpkgs-simple)

This example is functionally identical to the Nixpkgs example, with the exception
that it lets Nixpkgs fall back to `builtins.currentSystem`, which requires the
`--impure` command-line flag.

A real configuration would more likely list its systems explicitly, as the
Nixpkgs example does.
