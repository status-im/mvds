{

  /* This should match Nixpkgs commit in status-mobile. */
  source ? builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/e7603eba51f2c7820c0a182c6bbb351181caa8e7.tar.gz";
    sha256 = "sha256:0mwck8jyr74wh1b7g6nac1mxy6a0rkppz8n12andsffybsipz5jw";
  },
  pkgs ? import (source) {}
}:

pkgs.mkShell {
    name = "mvds-shell";

    buildInputs = with pkgs; [
      git jq which
      go_1_19 golangci-lint go-bindata
      protobuf3_21 protoc-gen-go
    ];
  }
