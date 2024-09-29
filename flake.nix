{
  description = "Leng, a fast dns proxy, built to black-hole internet advertisements and malware servers";


  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils, ... }:
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in {

        # Build & packaging
        ## use with `nix build`
        packages = rec {
          leng = pkgs.buildGoModule {
            inherit system;
            vendorHash = null;
            pname = "leng";
            version = "1.6.0";
            src = nixpkgs.lib.sources.cleanSource ./.;
            ldflags = [ "-s -w " ];
          };
          leng-container-image = pkgs.dockerTools.buildImage {
            name = "leng";
            created = "now";
            tag = "nix";
            copyToRoot = pkgs.buildEnv {
              name = "files";
              paths = [ leng pkgs.cacert ];
            };
            config.Entrypoint = [ "${leng}/bin/leng" ];
          };
          default = leng;
        };


        # Dev environment
        ## use with `nix develop`
        devShells = rec {
          # main development shell
          leng = with pkgs; mkShell {
            packages = [ fish go mdbook ];
            # Note that `shellHook` still uses bash syntax. This starts fish, then exists the bash shell when fish exits.
            shellHook = "fish && exit";
          };

          # shell with dependencies to build docs only
          ci-doc = with pkgs; mkShell {
            packages = [ mdbook mdbook-mermaid ];
          };

          default = leng;
        };


        # App
        ## use with `nix run`
        apps = rec {
          leng = flake-utils.lib.mkApp { drv = self.packages.${system}.leng; };
          default = leng;
        };

        checks = {
          inherit (self.packages.${system}) leng;
          metrics-api = pkgs.callPackage ./nixos-tests/metrics-api.nix { inherit self; };
          systemctl-start = pkgs.callPackage ./nixos-tests/systemctl-start.nix { inherit self; };
          custom-dns = pkgs.callPackage ./nixos-tests/custom-dns.nix { inherit self; };
        };

      }))
    //
    {
      nixosModules.default = { pkgs, lib, config, ... }:
        with lib;
        let
          cfg = config.services.leng;
          toml = pkgs.formats.toml { };
        in
        {
          ## interface
          options.services.leng = {
            enable = mkOption {
              type = types.bool;
              default = false;
            };
            enableSeaweedFsVolume = mkOption {
              type = types.bool;
              description = "Whether to make this nomad client capable of hosting a SeaweedFS volume";
            };
            package = mkOption {
              type = types.package;
              default = self.packages.${pkgs.system}.leng;
            };
            configuration = mkOption {
              type = toml.type;
              default = { };
              description = "Configuration as Nix attrSet";
              example = ''
                {
                  api = "127.0.0.1:8080";
                  metrics.enabled = true;
                  blocking.sourcesStore = "/var/lib/leng-sources";
                }
              '';
            };

          };

          ## implementation
          config = mkIf cfg.enable {
            environment = {
              etc."leng.toml".source = { blocking.sourcesStore = "/var/lib/leng-sources"; } // (toml.generate "leng.toml" cfg.configuration);
              systemPackages = [ cfg.package ];
            };

            systemd.services.leng = {
              description = "leng";
              wantedBy = [ "multi-user.target" ];
              wants = [ "network-online.target" ];
              after = [ "network-online.target" ];
              restartTriggers = [ config.environment.etc."leng.toml".source ];

              serviceConfig = {
                DynamicUser = true;
                ExecReload = "${pkgs.coreutils}/bin/kill -HUP $MAINPID";
                ExecStart = "${cfg.package}/bin/leng" + " --config=/etc/leng.toml";
                KillMode = "process";
                KillSignal = "SIGINT";
                Restart = "on-failure";
                RestartSec = 2;
                TasksMax = "infinity";
                StateDirectory = "leng-sources";
                AmbientCapabilities = "CAP_NET_BIND_SERVICE";
              };

              unitConfig = {
                StartLimitIntervalSec = 10;
                StartLimitBurst = 3;
              };
            };
            assertions = [
              {
                assertion = (cfg.configuration ? "blocking" && cfg.configuration.blocking ? "sourcesStore");
                message = ''
                  `services.leng.configuration.blocking.sourcesStore` should be set to a directory leng can write to, but it is not set.
                '';
              }
            ];
          };
        };
    };
}

