{
  description = "Nix Dependency Manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    systems.url = "github:nix-systems/default";

    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };

    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };

    pre-commit-hooks = {
      url = "github:cachix/git-hooks.nix";

      inputs = {
        flake-compat.follows = "flake-compat";
        nixpkgs.follows = "nixpkgs";
      };
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      pre-commit-hooks,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        name = "yae";
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ (import ./overlay.nix) ];
        };
      in
      {
        packages = {
          default = pkgs.${name};
          ${name} = self.packages.${system}.default;
        };

        apps = {
          default = {
            inherit (pkgs.${name}) meta;

            type = "app";
            program = "${self.packages.${system}.default}/bin/${name}";
          };

          ${name} = self.apps.${system}.default;
        };

        formatter = nixpkgs.legacyPackages."${system}".nixfmt-rfc-style;

        checks.pre-commit-check = pre-commit-hooks.lib.${system}.run {
          src = ./.;

          hooks = {
            deadnix.enable = true;
            flake-checker.enable = true;
            nixfmt-rfc-style.enable = true;
            statix.enable = true;
          };
        };

        devShells.default = nixpkgs.legacyPackages.${system}.mkShell {
          inherit (self.checks.${system}.pre-commit-check) shellHook;

          buildInputs = self.checks.${system}.pre-commit-check.enabledPackages ++ [
            pkgs.go
          ];
        };
      }
    )
    // {
      overlays.default = import ./overlay.nix;
    };
}
