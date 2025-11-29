{
  lib,
  stdenv,
  clangStdenv,
  stdenvAdapters,
  buildGoModule,
  musl,
  gitMinimal,
}:
let
  pname = "yae";
  version = "2025.11.29";
in
buildGoModule.override
  {
    stdenv = if stdenv.isDarwin then clangStdenv else stdenvAdapters.useMoldLinker clangStdenv;
  }
  {
    inherit pname version;
    src = lib.cleanSource ./.;
    vendorHash = "sha256-XQEB2vgiztbtLnc7BR4WTouPI+2NDQXXFUNidqmvbac=";
    buildInputs = if stdenv.isDarwin then [ ] else [ musl ];
    propagatedBuildInputs = [ gitMinimal ];

    ldflags = [
      "-s"
      "-w"
      "-X main.Version=${version}"
      "-X main.Commit=${version}"
    ]
    ++ (
      if stdenv.isDarwin then
        [ ]
      else
        [
          "-linkmode=external"
          "-extldflags=-static"
        ]
    );

    meta = with lib; {
      description = "Nix Dependency Manager";
      homepage = "https://github.com/Fuwn/${pname}";
      license = licenses.gpl3Only;
      maintainers = [ maintainers.Fuwn ];
      mainPackage = pname;
      platforms = platforms.unix;
    };
  }
