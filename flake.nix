{
  inputs = {

    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

    fenix = {
      url = "github:nix-community/fenix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

  };

  outputs =
    { self
    , nixpkgs
    , fenix
    , gitignore
    }:
    let
      inherit (gitignore.lib) gitignoreSource;

      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system:
        import nixpkgs {
          inherit system;
          overlays = [ self.overlay ];
        });

      chanspec = {
        date = "2022-01-15";
        channel = "nightly";
        sha256 = "kGKh+zzI1lFEbuYxGJ1uqm+sP8A/b66f+/zDICvQtuk="; # set zeros after modifying channel or date
      };
      rustChannel = p: (fenix.overlay p p).fenix.toolchainOf chanspec;

    in
    {

      overlay = final: prev: {

        globstar-ruby = with final;
          let
            analyzer-name = "ruby";
            pname = "globstar-${analyzer-name}";
            packageMeta = (lib.importTOML ./linters/ruby/Cargo.toml).package;
            rustPlatform = makeRustPlatform {
              inherit (rustChannel final) cargo rustc;
            };
          in
          rustPlatform.buildRustPackage {
            inherit pname;
            inherit (packageMeta) version;
            cargoBuildFlags = [ "-p" "${analyzer-name}" ];
            src = gitignoreSource ./.;
            cargoLock.lockFile = ./Cargo.lock;
          };

        globstar-dockerfile = with final;
          let
            analyzer-name = "dockerfile";
            pname = "globstar-${analyzer-name}";
            packageMeta = (lib.importTOML ./linters/dockerfile/Cargo.toml).package;
            rustPlatform = makeRustPlatform {
              inherit (rustChannel final) cargo rustc;
            };
          in
          rustPlatform.buildRustPackage {
            inherit pname;
            inherit (packageMeta) version;
            cargoBuildFlags = [ "-p" "${analyzer-name}" ];
            src = gitignoreSource ./.;
            cargoLock.lockFile = ./Cargo.lock;
          };

      };

      packages = forAllSystems (system: {
        inherit (nixpkgsFor."${system}") globstar-ruby globstar-dockerfile;
      });

      defaultPackage =
        forAllSystems (system: self.packages."${system}".globstar-dockerfile);

      devShell = forAllSystems (system:
        let
          pkgs = nixpkgsFor."${system}";
          toolchain = (rustChannel pkgs).withComponents [
            "rustc"
            "cargo"
            "rust-std"
            "rustfmt"
            "clippy"
            "rust-src"
          ];
          inherit (fenix.packages."${system}") rust-analyzer;
        in
        pkgs.mkShell {
          nativeBuildInputs = [
            pkgs.bacon
            pkgs.darwin.libiconv
            rust-analyzer
            toolchain

            pkgs.parallel
          ];
          RUST_LOG = "info";
          RUST_BACKTRACE = 1;
        });

    };
}
