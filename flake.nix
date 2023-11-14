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
          leng = pkgs.buildGo121Module {
            inherit system;
            vendorSha256 = "sha256-5dIZzqaw88lKuh1JHJurRZCPgrNzDHK/53bXKNGQBvQ=";
            pname = "leng";
            version = "0.0.1-test";
            src = ./.;
          };
          default = leng;
        };


        # Dev environment
        ## use with `nix develop`
        devShells = rec {
          # main development shell
          leng = with pkgs; mkShell {
            packages = [ fish go_1_21 mdbook ];
            # Note that `shellHook` still uses bash syntax. This starts fish, then exists the bash shell when fish exits.
            shellHook = "fish && exit";
          };

          # shell with dependencies to build docs only
          ci-doc = with pkgs; mkShell {
              packages = [ mdbook mdbook-d2 ];
          };

          default = leng;
        };


        # App
        ## use with `nix run`
        apps = rec {
          leng = flake-utils.lib.mkApp { drv = self.packages.${system}.leng; };
          default = leng;
        };

      }));
}
