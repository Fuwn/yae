{
  description = "Nix Dependency Manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-26.05-darwin";
    systems.url = "github:nix-systems/default";

    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        name = "yae";

        meta = with pkgs.lib; {
          description = "Nix Dependency Manager";
          homepage = "https://github.com/Fuwn/${name}";
          license = [
            licenses.mit
            licenses.asl20
          ];
          maintainers = [ maintainers.Fuwn ];
          mainProgram = name;
          platforms = platforms.unix;
        };

        yae = pkgs.buildGoModule rec {
          inherit meta;

          pname = name;
          version = "2025.11.29";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-XQEB2vgiztbtLnc7BR4WTouPI+2NDQXXFUNidqmvbac=";
          env.CGO_ENABLED = 0;
          nativeBuildInputs = [ pkgs.makeWrapper ];
          nativeCheckInputs = [
            pkgs.gitMinimal
          ];
          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${version}"
          ];

          postInstall = ''
            wrapProgram "$out/bin/yae" --prefix PATH : ${
              pkgs.lib.makeBinPath [
                pkgs.gitMinimal
                pkgs.nix
              ]
            }
          '';
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

        formatter = nixpkgs.legacyPackages."${system}".nixfmt;

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gitMinimal
            pkgs.nix
          ];
        };
      }
    );
}
