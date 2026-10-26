{
  outputs =
    { self }:
    let
      inherit (builtins.fromJSON (builtins.readFile "${self}/yae.json")) nixpkgs;

      pkgs =
        import
          (builtins.fetchTarball {
            inherit (nixpkgs) url sha256;
          })
          {
          };
    in
    {
      packages.${pkgs.stdenv.hostPlatform.system} = { inherit (pkgs) hello; };
    };
}
