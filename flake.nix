{
  description = "HupHop - multi-provider git portfolio manager";

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

      # Single source of truth: read from the VERSION file (trim trailing
      # newline). Same file the Go binary embeds and goreleaser overrides.
      version = builtins.replaceStrings [ "\n" ] [ "" ] (builtins.readFile ./VERSION);
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "hup";
          inherit version;
          src = self;
          # Update when Go dependencies change; nix reports the expected hash.
          vendorHash = "sha256-hN/YWgfUfYgtPbL8YwLcDMyldyJDTGyjTqjVom6sQds=";
          subPackages = [ "cmd/hup" ];
          # Tests run in the dedicated `gotest` check (which provides git).
          doCheck = false;
          ldflags = [ "-s" "-w" "-X main.version=${version}" ];
          meta = {
            description = "Multi-provider git portfolio manager";
            mainProgram = "hup";
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
            pkgs.goreleaser
            pkgs.gum
          ];
        };
      });

      checks = forAllSystems (pkgs: {
        # Build check.
        build = self.packages.${pkgs.system}.default;

        # Test + coverage gate: runs the whole suite (incl. e2e) via
        # scripts/coverage.sh and fails below >=70% overall / >=80% core.
        coverage = self.packages.${pkgs.system}.default.overrideAttrs (old: {
          pname = "hup-coverage";
          doCheck = true;
          # Sync/e2e tests shell out to git (and jj for the jj-clone path)
          # against local temp repos.
          nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.git pkgs.bash pkgs.jujutsu ];
          preCheck = ''
            export HOME=$(mktemp -d)
            git config --global user.email test@huphop.invalid
            git config --global user.name "huphop tests"
            git config --global init.defaultBranch main
          '';
          checkPhase = ''
            runHook preCheck
            bash scripts/coverage.sh
            runHook postCheck
          '';
        });
      });

      formatter = forAllSystems (pkgs: pkgs.nixpkgs-fmt);
    };
}
