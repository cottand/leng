{
  description = "Grimd, a fast dns proxy, built to black-hole internet advertisements and malware servers";


  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils, ... }:
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in {

        # Build & packaging
        ## use with `nix build`
        packages = rec {
          grimd = pkgs.buildGo121Module {
            inherit system;
            vendorSha256 = "sha256-5dIZzqaw88lKuh1JHJurRZCPgrNzDHK/53bXKNGQBvQ=";
            pname = "grimd";
            version = "0.0.1-test";
            src = ./.;
          };
          default = grimd;
        };


        # Dev environment
        ## use with `nix develop`
        devShells = rec {
          grimd = pkgs.mkShell {
            packages = [ pkgs.fish pkgs.go_1_21 ];
            # Note that `shellHook` still uses bash syntax. This starts fish, then exists the bash shell when fish exits.
            shellHook = "fish && exit";
          };
          default = grimd;
        };


        # App
        ## use with `nix run`
        apps = rec {
          grimd = flake-utils.lib.mkApp { drv = self.packages.${system}.grimd; };
          default = grimd;
        };

      }));
}
