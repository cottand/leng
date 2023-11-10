{
  description = "Grimd, a fast dns proxy, built to black-hole internet advertisements and malware servers";


  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      build = { system, vendorSha256 }:
        let pkgs = import nixpkgs { inherit system; }; in pkgs.buildGo121Module {
          inherit vendorSha256;
          pname = "grimd";
          version = "0.0.1-test";
          src = ./.;
        };
    in
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in rec {

        ## Build & packaging
        packages.grimd = build {
          inherit system;
          vendorSha256 = "sha256-5dIZzqaw88lKuh1JHJurRZCPgrNzDHK/53bXKNGQBvQ=";
        };

        defaultPackage = packages.grimd;


        ## Dev environment 
        devShells = rec {
          grimd = pkgs.mkShell {
            packages = [ pkgs.fish pkgs.go_1_21 ];
            # Note that `shellHook` still uses bash syntax. This starts fish, then exists the bash shell when fish exits.
            shellHook = ''
              fish && exit
            '';
          };
          default = grimd;
        };

        ## App
        apps = rec {
          grimd = flake-utils.lib.mkApp { drv = self.packages.${system}.grimd; };
          default = grimd;
        };
      }));
}
