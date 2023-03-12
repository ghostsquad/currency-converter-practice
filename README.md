<h1 align="center">
  <br>
  <a href="http://github.com/ghostsquad/currency-converter-practice"><img src="./docs/assets/exchange.png" alt="github.com/ghostsquad/s3-file-explorer" width="200px" /></a>
  <br>
  Currency Converter Practice Project
  <br>
</h1>

<p align="center">
  <a href="#introduction">Introduction</a> •
  <a href="#getting-started">Getting Started</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#roadmap">Roadmap</a>
</p>

## Introduction

A simple Golang webapp to get currency exchange information.

## Getting Started

### One-Time
```shell
# Ensure you have asdf installed
# brew install asdf

# TODO check that all tools have default plugins available
asdf plugin-add task https://github.com/particledecay/asdf-task.git
asdf plugin-add tanka
asdf plugin-add k3d

asdf install

# TODO add this to setup task (extra credit: cross platform?)
# this is needed for sponge https://linux.die.net/man/1/sponge
brew install moreutils

# Install jsonnet external libraries
# TODO ensure this is a dependency on all jsonnet tasks
task jb:install
```

```shell
task run
```

In a separate shell

```shell
task http:metrics
```

### Deploying to K3D

```shell
task k3d:up
task k3d:registry:up

# NOTE: registry up prints a message to update /etc/hosts
# This is not done automatically, as that is a system-wide change 
# that I feel is something the user should handle
# TODO investigate using dnsmasq or something else to prevent this requirement

task publish:k3d
task tk:apply -- environments/default
```

### Configuration

Configuration is done via environment variables. Standard AWS SDK Environment variables supported, as well as OIDC/EC2 authentication methods.

| Variable       | Required | Default | Description                                                    |
|----------------|----------|---------|----------------------------------------------------------------|
| `APP_TBD`      | no       | ``      | Just an example unused environment variable                    |
| `PORT`         | no       | `8080`  | The listen port                                                |
| `BIND_ADDRESS` | no       |         | Configured to listen on 127.0.0.1, this may not work in Docker |

## Roadmap

- [ ] Add Kubernetes Manifests (ideally this is put in a separate repo)
- [ ] Add K3D Support
- [ ] Add Github Actions CI/CD
- [ ] See TODO 🔍 comments all around... 

### Known Issues

1. There's a slew of app optimizations still left, such as separating `/ping` and `/metrics` endpoints onto different listeners (separate for application endpoints), make logging configurable between text/color and JSON and some other things. This was my first time using Gin.
2. I didn't configure test coverage or pay attention to that much. Just wrote a few small unit tests and a sanity-check end2end test (just verifies that the app runs and I can /ping it) via docker-compose.
