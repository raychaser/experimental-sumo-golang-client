# WARNING

This is an initial **EXPERIMENTAL** attempt at creating both
Sumo Logic API bindings for Golang as well as an actual CLI
to interact with the Sumo Logic service.

Please consider this "my first Go project"––the author is NOT
a professional Go programmer. However, an attempt was made to
follow best practices for building API and CLI clients.

At this point this code is not meant for production use. This
code is **NOT** an official product of Sumo Logic.


# What Is This?

-  [go-sumologic](go-sumologic/README.md) API bindings for the
   Sumo Logic service
-  [Sumo CLI](sumo-cli/README.md) A Sumo Logic CLI client based
   on the bindings provided by `go-sumologic`


# Build & Run

This has most recently been tested with Go `go1.16.3`.

On a Mac, one way to get the Go toolchain is to use Homebrew:

```sh
brew install golang
```

On Mac, you will also need `coreutils` (for `realpath`, which
is used in the build script):

```sh
brew install coreutils
```

Basic instructions to build:

```sh
git clone https://github.com/...
cd ...
./build-all.sh
# Binaries for all platforms will be in out/
tree out/
```

_TODO Shoud probably create proper makefiles..._

