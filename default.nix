let
  inherit (builtins)
    currentSystem
    fromJSON
    readFile
    ;

  flakeLock = fromJSON (readFile ./flake.lock);
  fromFlakeLock =
    node: with flakeLock.nodes.${node}.locked; {
      inherit rev;
      outPath = fetchTarball {
        url = "https://github.com/${owner}/${repo}/archive/${rev}.tar.gz";
        sha256 = narHash;
      };
    };
in
{
  system ? currentSystem,
  pkgs ? import (fromFlakeLock "nixpkgs") {
    hostPlatform = system;
  },
}:
pkgs.callPackage ./package.nix { }
