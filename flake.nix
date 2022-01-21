{
  inputs = {

    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

    fenix = {
      url = "github:nix-community/fenix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

  };

  outputs =
    { self
    , nixpkgs
    , fenix
    }:
    let
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

      overlay = final: prev: { };

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
