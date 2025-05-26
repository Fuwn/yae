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
        inherit (pkgs.stdenv) isDarwin;

        pkgs = import nixpkgs { inherit system; };
        name = "yae";

        meta = with pkgs.lib; {
          description = "Nix Dependency Manager";
          homepage = "https://github.com/Fuwn/${name}";
          license = licenses.gpl3Only;
          maintainers = [ maintainers.Fuwn ];
          mainPackage = name;
          platforms = platforms.unix;
        };

        yae =
          pkgs.buildGo123Module.override
            {
              stdenv = if isDarwin then pkgs.clangStdenv else pkgs.stdenvAdapters.useMoldLinker pkgs.clangStdenv;
            }
            rec {
              inherit meta;

              pname = name;
              version = "2025.05.25";
              src = pkgs.lib.cleanSource ./.;
              vendorHash = "sha256-XQEB2vgiztbtLnc7BR4WTouPI+2NDQXXFUNidqmvbac=";
              buildInputs = if isDarwin then [ ] else [ pkgs.musl ];
              propagatedBuildInputs = [ pkgs.gitMinimal ];

              ldflags =
                [
                  "-s"
                  "-w"
                  "-X main.Version=${version}"
                  "-X main.Commit=${version}"
                ]
                ++ (
                  if isDarwin then
                    [ ]
                  else
                    [
                      "-linkmode=external"
                      "-extldflags=-static"
                    ]
                );
            };
      in
      {
        packages = {
          default = yae;
          ${name} = self.packages.${system}.default;
        };

        apps = {
          default = {
            inherit meta;

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
            pkgs.go_1_23
          ];
        };
      }
    );
}
