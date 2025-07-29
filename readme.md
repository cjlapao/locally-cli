# Locally


locally is a command line tool design to help spin up a local environment including the infrastructure, the concept is easy, have configuration files created by each team that can be shareable and reproducible from machine to machine and deploy the bare minimum infrastructure.

## How to install

locally is not installed, it is just an executable with example/template configuration bundle so you can just download the latest release, unzip it to a folder and put it in the environment path.  


## Troubleshoot

We have a troubleshoot guide [here](./docs/troubleshooting.md) where we place the most common issues found by people, this will be constantly updated, so please be sure to read it before you ask questions.  
If you do not find an answer there you can use the locally channel to ask a question to the team [here](https://teams.microsoft.com/l/channel/19%3a98b5d070649f442ab23b247ec5858e16%40thread.skype/locally%2520-%2520Also%2520called%2520Locally?groupId=cd5ee759-4aef-4928-95f4-b8c658c5d0db&tenantId=e5208e76-dd12-47f0-9541-c9b45afaffe6).  


## Building locally locally

locally is written in Go and uses VSCODE to easily debug, you still need a few tools if you do not have in case you want to build it from source, or just debug it.

### Getting and installing go onto your PC

Download the latest Go from [here](https://go.dev/dl/), choose your operating system and then run. Note: if you are running Mac or Linux this needs to be unziped to a folder.  

Once this is done you can quickly test it by typing ```go version```

### Visual Studio Code with GO

There is a good setup guide in [here](https://code.visualstudio.com/docs/languages/go) this will use the extensions provided by google and allow intellisense in vscode
this is the [extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)

### How to build

The project includes a comprehensive Makefile for cross-platform building. All builds include SQLite support by default (CGO enabled). Here are the available build options:

#### Quick Build (Current Platform)
```bash
# Build for current platform
make build

# Build API only
make api-build
```

#### Cross-Platform Building

**Build for specific platform:**
```bash
# Linux (amd64)
make build-linux

# Linux (arm64)
make build-linux-arm64

# macOS (amd64)
make build-macos

# macOS (arm64/M1)
make build-macos-arm64

# Windows (amd64) - Requires Windows C compiler for cross-compilation
make build-windows

# Windows (arm64) - Requires Windows C compiler for cross-compilation
make build-windows-arm64
```

**Build for all platforms:**
```bash
# Build for all major platforms
make build-all
```

**Custom platform build:**
```bash
# Build for specific OS/arch combination
GOOS=linux GOARCH=amd64 make build-cross
GOOS=darwin GOARCH=arm64 make build-cross
GOOS=windows GOARCH=amd64 make build-cross
```

> **Note:** Cross-compiling with CGO (required for SQLite) has limitations:
> - **Same platform, different architecture**: Requires target architecture's C compiler
>   - Linux amd64 → Linux arm64: Requires ARM64 GCC
>   - macOS amd64 → macOS arm64: Requires ARM64 Clang
> - **Cross-platform**: Requires target platform's C compiler
>   - Linux → Windows: Requires MinGW-w64
>   - Linux → macOS: Requires Xcode command line tools
>   - macOS → Linux: Requires GCC
>   - macOS → Windows: Requires MinGW-w64
> 
> **What works out of the box:**
> - Building for the current platform and architecture
> - Building for the same OS but different architecture (if C compiler is available)
> 
> For reliable cross-platform builds, use native builds on each platform or CI/CD pipelines.

#### Release Builds

**Build release binaries:**
```bash
# All platforms (with SQLite support)
make release
```

#### Build Parameters

You can customize builds with these parameters:

- `GOOS`: Target operating system (linux, darwin, windows)
- `GOARCH`: Target architecture (amd64, arm64, 386)
- `CGO_ENABLED`: Enable CGO for SQLite (default: 1)

**Examples:**
```bash
# Build for Linux ARM64
GOOS=linux GOARCH=arm64 make build-cross

# Build for Windows AMD64
GOOS=windows GOARCH=amd64 make build-cross

# Build without CGO (not recommended - SQLite won't work)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make build-cross
```

#### Manual Go Build

If you prefer to use Go directly:

```bash
# Build for current platform (with SQLite)
CGO_ENABLED=1 go build -o locally-api ./cmd/api

# Build for specific platform (with SQLite)
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o locally-api-linux ./cmd/api
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o locally-api-macos ./cmd/api
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o locally-api-windows.exe ./cmd/api
```

### Docker Registry Management

The project includes Docker registry targets for building and pushing to `dcr.carloslapao.com/locally/locally-api`:

#### Docker Registry Commands

```bash
# Login to registry
make docker-login

# Build and push to registry
make docker-push

# Build and push with specific tag
DOCKER_TAG=v1.0.0 make docker-push

# Build and push with latest tag
make docker-push-latest

# Build and push with version from VERSION file
make docker-push-version

# Pull from registry
make docker-pull

# Tag local image for registry
make docker-tag

# Build, tag, and push in one command
make docker-build-and-push

# Build and push with latest tag
make docker-build-and-push-latest

# Build and push with version tag
make docker-build-and-push-version
```

#### Registry Configuration

The registry configuration can be customized:

```bash
# Use different registry
DOCKER_REGISTRY=my-registry.com make docker-push

# Use different namespace
DOCKER_NAMESPACE=my-namespace make docker-push

# Use different image name
DOCKER_IMAGE=my-api make docker-push

# Use different tag
DOCKER_TAG=v2.0.0 make docker-push
```

#### Running from Registry

The Docker run scripts automatically use the registry image:

```bash
# Run with latest tag
make docker-run

# Run with specific tag
DOCKER_TAG=v1.0.0 make docker-run

# Run with custom registry
DOCKER_REGISTRY=my-registry.com make docker-run
```
