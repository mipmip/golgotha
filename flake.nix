{
  description = "Skull2 - multi-provider git portfolio manager";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      # Plain nix multi-arch (no flake-utils).
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f:
        nixpkgs.lib.genAttrs supportedSystems (system: f (import nixpkgs { inherit system; }));

      version = "0.0.0-dev";
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "skull2";
          inherit version;
          src = self;
          # Update when Go dependencies change; nix reports the expected hash.
          vendorHash = "sha256-g+yaVIx4jxpAQ/+WrGKxhVeliYx7nLQe/zsGpxV4Fn4=";
          subPackages = [ "cmd/skull2" ];
          ldflags = [ "-s" "-w" "-X main.version=${version}" ];
          meta = {
            description = "Multi-provider git portfolio manager";
            mainProgram = "skull2";
          };
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gopls
            pkgs.gotools
            pkgs.golangci-lint
            pkgs.git
            pkgs.jujutsu
          ];
        };
      });

      checks = forAllSystems (pkgs: {
        # Build check.
        build = self.packages.${pkgs.system}.default;

        # Test check. The coverage gate (>=70% overall, >=80% core) is wired in
        # here during milestone "05 Testing, e2e & coverage".
        gotest = self.packages.${pkgs.system}.default.overrideAttrs (_: {
          pname = "skull2-tests";
          doCheck = true;
        });
      });

      formatter = forAllSystems (pkgs: pkgs.nixpkgs-fmt);
    };
}
