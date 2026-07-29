{ pkgs }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
    gcc
    valkey
  ];
}
