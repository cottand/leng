{ self, pkgs, home-manager, ... }:
let
  nixpkgs = self.inputs.nixpkgs;
in
(nixpkgs.lib.nixos.runTest {
  hostPkgs = pkgs;
  defaults.documentation.enable = false;
  node.specialArgs = { inherit self; };

  name = "leng-custom-dns";

  nodes = {
    server = { config, pkgs, ... }: {
      imports = [ self.nixosModules.default ];
      # Open the default port for `postgrest` in the firewall
      networking.firewall.allowedUDPPorts = [ 53 ];

      services.leng.enable = true;
      services.leng.configuration = {
        blocking.sourcesStore = "/tmp";
        customdnsrecords = [
          "example.com    IN A   1.2.3.4"
        ];
      };
    };

    client = { pkgs, ... }: {
      environment.systemPackages = [ pkgs.dig ];
    };
  };

  testScript =
    ''
      start_all()

      server.wait_for_unit("leng", timeout=10)
      server.wait_for_open_port(53)

      client.succeed(
        "dig @server example.com | grep -o '1.2.3.4' "
      )
    '';

}).config.result
