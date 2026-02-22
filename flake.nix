{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixpkgs-master.url = "github:NixOS/nixpkgs/master";
  };
  outputs = {
    self,
    nixpkgs,
    nixpkgs-master,
  }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs {inherit system;};
    master-pkgs = import nixpkgs-master {inherit system;};
  in {
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = [
        master-pkgs.go_1_26
        master-pkgs.golangci-lint
        pkgs.gnumake
        pkgs.zig
      ];
      shellHook = ''
        go env -w GOPATH=$HOME/.local/share/go
        export PATH="$HOME/.local/bin:$PATH"
        export PATH="$HOME/.local/share/go/bin:$PATH"
        export ZIG_GLOBAL_CACHE_DIR="/tmp"
      '';
    };
  };
}
