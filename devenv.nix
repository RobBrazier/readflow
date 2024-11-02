{ pkgs, lib, config, inputs, ... }:

{
  packages = [ pkgs.git pkgs.go-task pkgs.nodejs_22 pkgs.goreleaser ];

  languages.go.enable = true;

  enterTest = ''
    go test ./...
  '';

  pre-commit.hooks = {
    govet.enable = true;
    gofmt.enable = true;
  };

  # See full reference at https://devenv.sh/reference/options/
}
