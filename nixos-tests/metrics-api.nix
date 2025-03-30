{ self, pkgs, home-manager, ... }:
let
  nixpkgs = self.inputs.nixpkgs;
  httpPort = 9243;
in
(nixpkgs.lib.nixos.runTest {
  hostPkgs = pkgs;
  defaults.documentation.enable = false;
  node.specialArgs = { inherit self; };

  name = "leng-systemctl-start";

  nodes = {
    server = { config, pkgs, ... }: {
      imports = [ self.nixosModules.default ];
      # Open the default port for `postgrest` in the firewall
      networking.firewall.allowedTCPPorts = [ httpPort ];

      services.leng.enable = true;
      services.leng.configuration = {
        metrics.enabled = true;
        api = "0.0.0.0:${toString httpPort}";
        blocking.sourcesStore = "/tmp";
      };
    };

    client = { };
  };

  testScript =
    ''
      start_all()

      server.wait_for_unit("leng", timeout=10)
      server.wait_for_open_port(${toString httpPort})

      actual = client.succeed(
        "curl http://server:${toString httpPort}/metrics | grep -o 'go_gc_duration_seconds' "
      )
    '';

}).config.result
