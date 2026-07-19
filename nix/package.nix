{ pkgs }:

pkgs.buildGoModule {
  pname = "hellbot";
  version = "0.1.0";
  src = ./..;

  vendorHash = null;

  nativeBuildInputs = with pkgs; [ gcc ];

  env.CGO_ENABLED = "1";
}
