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
          vendorHash = "sha256-TSpb8oiLdVlBpUCAlRXx+LQt8ZcZZslId02hP3qRRlA=";
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

          preBuild = ''
            export HOME="$TMPDIR"
          '';

          checkPhase = ''
            runHook preCheck
            go test ./...
            go vet ./...
            test -z "$(gofmt -l *.go internal)"
            runHook postCheck
          '';

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

        checks = {
          package = yae;
          runtime = pkgs.runCommand "yae-runtime-check" { nativeBuildInputs = [ pkgs.go ]; } ''
            export HOME="$TMPDIR" GOCACHE="$TMPDIR/go-cache" CGO_ENABLED=0
            export YAE_TEST_BINARY=${yae}/bin/yae YAE_TEST_GIT=${pkgs.gitMinimal}/bin/git
            go test -v ${./runtime_test.go} -run '^TestPackagedRuntime$'
            touch "$out"
          '';

          nix =
            pkgs.runCommand "yae-nix-check"
              {
                nativeBuildInputs = [
                  pkgs.deadnix
                  pkgs.flake-checker
                  pkgs.nixfmt
                  pkgs.statix
                  pkgs.actionlint
                ];
              }
              ''
                export HOME="$TMPDIR"

                cd ${self}
                deadnix --fail .
                flake-checker -f
                nixfmt --check flake.nix examples/*/flake.nix
                statix check .
                actionlint .github/workflows/check.yml
                touch "$out"
              '';
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gitMinimal
            pkgs.nix
            pkgs.deadnix
            pkgs.flake-checker
            pkgs.nixfmt
            pkgs.statix
            pkgs.actionlint
          ];
        };
      }
    );
}
