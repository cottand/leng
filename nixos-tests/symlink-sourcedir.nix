{ self, pkgs, home-manager, ... }:
let
  nixpkgs = self.inputs.nixpkgs;
in
(nixpkgs.lib.nixos.runTest {
  hostPkgs = pkgs;
  defaults.documentation.enable = false;
  node.specialArgs = { inherit self; };

  name = "leng-symlink-sourcedir";

  nodes = {
    server = { config, pkgs, ... }: {
      imports = [ self.nixosModules.default ];
      services.leng.enable = true;
      services.leng.configuration = {
        blocking.sourcesStore = "/var/lib/leng-sources";
        blocking.sourcedirs = [];
      };
    };
  };

  testScript =
    ''
      start_all()

      server.wait_for_unit("leng", timeout=10)
    '';

}).config.result
