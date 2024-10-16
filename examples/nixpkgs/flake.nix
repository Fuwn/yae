{
  outputs =
    { self }:
    let
      nixpkgs = (builtins.fromJSON (builtins.readFile "${self}/yae.json")).nixpkgs;

      systemsFlakeExposed = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "armv6l-linux"
        "armv7l-linux"
        "i686-linux"
        "aarch64-darwin"
        "powerpc64le-linux"
        "riscv64-linux"
        "x86_64-freebsd"
      ];

      forEachSystem =
        systems: action:
        builtins.listToAttrs (
          map (system: {
            name = system;
            value = action system;
          }) systems
        );
    in
    {
      packages = (forEachSystem systemsFlakeExposed) (
        system:
        let
          pkgs =
            import
              (builtins.fetchTarball {
                inherit (nixpkgs) url sha256;
              })
              {
                inherit system;
              };
        in
        {
          hello = pkgs.hello;
        }
      );
    };
}
