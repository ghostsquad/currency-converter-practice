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

A simple Golang webapp to get currency exchange information, purposefully overengineered to showcase all the different components of a production-ready application, including:

- Local Testing
- Container Testing (via `docker-compose`)
- Local K8s Testing (via `k3d`)
- TDD
- Metrics
- Readiness/Liveness Probes (/ping)
- Using Task (an alternative to make)
- Layers (a custom jsonnet-based system for sharing configuration across process boundaries)
- Leaning into Git (automatically, such as using the SHA for publishing images)
- and more!

📌 **Note**

The scope of this project increased exponentially as I was working on it, and I had to decide how far to actually take certain things...
Otherwise I'd simply be working on this forever. 

The goal is to continuously refine this project and bake more and more "defaults" into reusable libraries and configuration "Layers".

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

### Testing Via Docker-Compose

```shell
task test:integration
```

### Deploying to K3D

```shell
# TODO combine these
task k3d:create

# NOTE: registry up prints a message to update /etc/hosts
# This is not done automatically, as that is a system-wide change 
# that I feel is something the user should handle
# TODO investigate using dnsmasq or something else to prevent this requirement

task publish:k3d
task tk:apply -- environments/default
task test:integration:k3d
```

### Configuration

Configuration is done via environment variables.

| Variable       | Required | Default | Description                                                    |
|----------------|----------|---------|----------------------------------------------------------------|
| `APP_TBD`      | no       | ``      | Just an example unused environment variable                    |
| `PORT`         | no       | `8080`  | The listen port                                                |
| `BIND_ADDRESS` | no       |         | Configured to listen on 127.0.0.1, this may not work in Docker |

## Roadmap

- [x] Add Kubernetes Manifests (ideally this is put in a separate repo)
- [x] Add K3D Support
- [ ] Add Github Actions CI/CD
- [ ] See TODO 🔍 comments all around... 

### Known Issues/Notes

1. There's a slew of app optimizations still left, such as separating `/ping` and `/metrics` endpoints onto different listeners (separate for application endpoints), make logging configurable between text/color and JSON and some other things. This was my first time using Gin.
2. I didn't configure test coverage or pay attention to that much. Just wrote a few small unit tests and a sanity-check end2end test (just verifies that the app runs, I can /ping it, and a small assertion of the expected output) via docker-compose and k3d.
3. go and jsonnet/jb/tanka fight over the `vendor` directory. This is somewhat fixed with `-mod=readonly` or `-mod=mod` as an explicit flag to go commands, or `GOFLAGS=-mod=mod`, however, Jetbrains GOLAND still has problems if the `vendor` directory exists and doesn't have go files in it. See https://youtrack.jetbrains.com/issue/GO-10952/No-way-to-disable-vendoring-for-Sync-Dependencies-quick-fix. To get Goland to not freak out, `rm -rfd vendor`. This of course breaks all the Jsonnet, so when you are ready to do anything else, run `task jb:install`. Another possible way to fix this is if tanka supported other vendor directories. See https://github.com/grafana/tanka/issues/356 and https://github.com/grafana/tanka/issues/820.  
4. When making changes to `taskfile.jsonnet` there's a really painful loop of:
   1. copy the vars block (because of the issue with ordering in a map). 
   2. run `task taskfile:gen`
   3. paste the vars block (optionally: if changes to vars were made, make sure those are captured)
   4. run `task <blah>` (whatever it is what I was working on at the time)

   Solution: When I first found Taskfile, I really liked it, but it turns out to not really be as flexible or intuitive as I prefer. I started a project called [fngo](https://fngo.dev/) as a replacement for Task, but I'm still in the early stages of really understanding what I want, and how I want to represent it.
5. Layers is still a WIP. Because of that, it was a bit painful to update 2 repos when I found some good patterns. But that's also the point, is that those patterns are now available for future projects without much work.
6. Jsonnet + Taskfile is weird... It kind of reminds me of Terraform + Jsonnet, in that there are some values that are known at runtime and other values that are known at compile-time (jsonnet natively derived/defined values). Figuring out the best way to present those 2 types and keep things orderly and making sense will be a huge step towards making Layers and/or fngo really nice to use.
