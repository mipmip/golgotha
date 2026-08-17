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
          vendorHash = "sha256-aJllcMJduoi8VBWMJWsxm8swXtNonYZzX8etmNZePzc=";
          subPackages = [ "cmd/skull2" ];
          # Tests run in the dedicated `gotest` check (which provides git).
          doCheck = false;
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
        gotest = self.packages.${pkgs.system}.default.overrideAttrs (old: {
          pname = "skull2-tests";
          doCheck = true;
          # Sync/e2e tests shell out to git against local temp repos.
          nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.git ];
          preCheck = ''
            export HOME=$(mktemp -d)
            git config --global user.email test@skull2.invalid
            git config --global user.name "skull2 tests"
            git config --global init.defaultBranch main
          '';
        });
      });

      formatter = forAllSystems (pkgs: pkgs.nixpkgs-fmt);
    };
}
