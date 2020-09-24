{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [ stdenv go glibc.static ];
  CFLAGS = "-I${pkgs.glibc.dev}/include";
  LDFLAGS = "-L${pkgs.glibc}/lib";
}
