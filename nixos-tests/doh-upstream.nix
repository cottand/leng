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
      networking.firewall.allowedTCPPorts = [ 80 ];

      services.leng.enable = true;
      services.leng.configuration = {
        blocking.sourcesStore = "/tmp";
        customdnsrecords = [ ];
        upstream.DoH = "";
        DnsOverHttpServer.enabled = true;
      };
    };

    client = { pkgs, ... }: {
      environment.systemPackages = [ pkgs.dig pkgs.curl ];
    };
  };

  testScript =
    ''
      start_all()

      server.wait_for_unit("leng", timeout=10)
      server.wait_for_open_port(80, timeout=10)

      client.succeed(
        'curl -vH "accept: application/dns-json" "http://server/dns-query?dns=AAABAAABAAAAAAAAA3d3dwdleGFtcGxlA2NvbQAAAQAB"',
        timeout=10,
      )
    '';

}).config.result
