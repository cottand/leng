{ pkgs ? import (builtins.fetchTarball "https://api.github.com/repos/nixos/nixpkgs/tarball/nixos-unstable") {} }:
  pkgs.mkShell {
    nativeBuildInputs = with pkgs; [ go_1_21 ];
}
