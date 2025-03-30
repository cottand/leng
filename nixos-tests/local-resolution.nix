{ self, pkgs, home-manager, ... }:
let
  nixpkgs = self.inputs.nixpkgs;
in
(nixpkgs.lib.nixos.runTest {
  hostPkgs = pkgs;
  defaults.documentation.enable = false;
  node.specialArgs = { inherit self; };

  name = "leng-local-resolution";

  nodes = {
    server = { config, pkgs, ... }: {
      imports = [ self.nixosModules.default ];
      environment.systemPackages = [ pkgs.dig ];
      networking.firewall.allowedUDPPorts = [ 53 ];
      networking.nameservers = [ "127.0.0.1" ];

      services.leng.enable = true;
      services.leng.configuration = {
        blocking.sourcesStore = "/tmp";
        customdnsrecords = [
          "example.com    IN A   1.2.3.4"
        ];
      };
    };
  };

  testScript =
    ''
      start_all()

      server.wait_for_unit("leng", timeout=10)
      server.wait_for_open_port(53)

      server.succeed(
        "dig example.com"
      )
    '';

}).config.result
