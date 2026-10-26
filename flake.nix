{
  description = "Nix Dependency Manager";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-26.05-darwin";

  # Tsutsumi calls these outputs directly with its own arguments and a plain
  # store path as self, so do not add inputs or read flake metadata from self.
  outputs =
    { self, nixpkgs, ... }:
    let
      forEachSystem =
        function:
        nixpkgs.lib.genAttrs [
          "aarch64-darwin"
          "aarch64-linux"
          "x86_64-darwin"
          "x86_64-linux"
        ] (system: function nixpkgs.legacyPackages.${system});

      yaeFor =
        pkgs:
        pkgs.buildGoModule rec {
          pname = "yae";
          version = "2025.11.29";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-CAQbD3gfKCUoB7nT7i/4J2ymkua+LmFT0CDAqoHX4AA=";
          env.CGO_ENABLED = 0;
          nativeBuildInputs = [ pkgs.makeWrapper ];
          nativeCheckInputs = [ pkgs.gitMinimal ];
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

          meta = {
            description = "Nix Dependency Manager";
            homepage = "https://github.com/Fuwn/yae";
            license = [
              pkgs.lib.licenses.mit
              pkgs.lib.licenses.asl20
            ];
            maintainers = [ pkgs.lib.maintainers.Fuwn ];
            mainProgram = "yae";
            platforms = pkgs.lib.platforms.unix;
          };
        };

      nixCheckTools = pkgs: [
        pkgs.deadnix
        pkgs.nixfmt
        pkgs.statix
        pkgs.actionlint
      ];
    in
    {
      packages = forEachSystem (
        pkgs:
        let
          yae = yaeFor pkgs;
        in
        {
          inherit yae;

          default = yae;
        }
      );

      formatter = forEachSystem (pkgs: pkgs.nixfmt);

      checks = forEachSystem (
        pkgs:
        let
          yae = yaeFor pkgs;
        in
        {
          package = yae;

          runtime = pkgs.runCommand "yae-runtime-check" { nativeBuildInputs = [ pkgs.gitMinimal ]; } ''
            export HOME="$TMPDIR" NIX_REMOTE="local?root=$TMPDIR/nix" NIX_CONF_DIR="$TMPDIR/config"
            export NIX_USER_CONF_FILES= NIX_CONFIG="build-users-group =" GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null

            sources="$TMPDIR/sources.json"
            repository="$TMPDIR/repository"
            sha256=1n8lvnnjp3g8a2bj9msjxibi7sjfq9570vyyjr6pl6wi4xqxrbdx
            hash=sha256-va3ccSeRG3pNlt5vcErCTuoTV+xS1ySXUOiNK63dFNk=

            # An empty PATH proves that the wrapper supplies Git and Nix.
            yae() { PATH= ${yae}/bin/yae --sources "$sources" "$@"; }

            yae init
            printf 'first payload' > "$TMPDIR/payload"
            yae add --type binary --unpack=false sample "file://$TMPDIR/payload"
            grep -qF "\"hash\": \"$hash\"" "$sources"

            git init --quiet --bare "$repository"
            commit=$(git -C "$repository" -c user.name=Yae -c user.email=yae@example.test commit-tree "$(git -C "$repository" mktree < /dev/null)" -m fixture)
            git -C "$repository" update-ref refs/tags/v10 "$commit"

            export GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0="url.file://$repository.insteadOf" GIT_CONFIG_VALUE_0=https://example.test/owner/repository
            printf '{"current": {"url": "%s", "url_template": "%s", "version": "v10", "type": "git", "unpack": false, "sha256": "%s", "hash": "%s"}}' \
              https://example.test/owner/repository/archive/v10 \
              'https://example.test/owner/repository/archive/{version}' \
              "$sha256" "$hash" > "$sources"
            test -z "$(yae update --output-updated-list)"
            touch "$out"
          '';

          nix = pkgs.runCommand "yae-nix-check" { nativeBuildInputs = nixCheckTools pkgs; } ''
            export HOME="$TMPDIR"

            cd ${self}
            deadnix --fail .
            nixfmt --check flake.nix examples/*/flake.nix
            statix check .
            actionlint .github/workflows/check.yml
            touch "$out"
          '';
        }
      );

      devShells = forEachSystem (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gitMinimal
            pkgs.nix
          ]
          ++ nixCheckTools pkgs;
        };
      });
    };
}
