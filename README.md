# DDD & Hexagonal Architecture with Standard Go Project Layout

This Go module takes part of the [mmw](https://github.com/pivaldi/mmw) project that demonstrates the implementation of the [Go Modular Monolith White Paper](https://github.com/pivaldi/go-modular-monolith-white-paper).

This project is not usable independently of [mmw-auth](https://github.com/pivaldi/mmw-auth); the best way to test this project is to use directly the [Monolith Modular Worskpace](https://github.com/pivaldi/mmw).

## Overview

This repository includes a **working example implementation** that demonstrates how to apply Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns using the [Standard Go Project Layout](https://github.com/golang-standards/project-layout/releases/latest)).

As is, this project takes part of [the `mmw` project](https://github.com/pivaldi/mmw) and should be use for now from this Monolith Modular Workspace.

### What's Included

The Todo API example provides:

- **Complete CRUD API** for managing todo items
- **Dual Protocol Support** - HTTP and gRPC from single protobuf definitions using [Buf Connect](https://connect.build)
- **In-process/Network calls switch** - demonstrate the interaction with an authentication service in-proccess or through a gRPC connection.
- **Domain-Driven Design** - Rich domain model with aggregates, value objects, and domain events
- **Hexagonal Architecture** - Clear separation between domain, application, and infrastructure layers
- **PostgreSQL Persistence** - Repository pattern with database migrations
- **Comprehensive Testing** - Unit, integration, and API tests demonstrating testing strategies for each layer
- **Production-Ready Structure** - Docker support, CI/CD configuration, and operational tooling
- **Angular Web Application** - Modern, responsive frontend for managing todos (`/web/todoapp`)

### Quick Start

```bash
# Bootstrap the project: install tools APP_ENV aware
./configure

# Start the project
mise run db:up
mise run db:migrate:up
cd web/todoapp/ && npm install && npm start
```

> Mise provides automatic tool version management and better cross-platform support. See [`docs/MISE.md`](docs/MISE.md) for details.

One the database is migrated you can use Zellij to bootstrap the project with the unique command `mise run zellij:dev`

### Architecture Highlights

The implementation demonstrates:

- **Domain Layer** (`/internal/domain/todo`) - Pure business logic with zero external dependencies
- **Application Layer** (`/internal/application`) - Use case orchestration and transaction management
- **Ports** (`/internal/application/ports`) - Interface definitions for dependency inversion
- **Adapters** (`/internal/adapters`) - Concrete implementations (HTTP/gRPC handlers, PostgreSQL repositories)
- **API Definitions** (`/api`) - Protobuf schemas with Buf for code generation
- **Web Frontend** (`/web/todoapp`) - Angular application with CRUD operations, filtering, and real-time updates

For detailed architecture documentation, see [`docs/plans/2026-02-03-todo-api-design.md`](docs/plans/2026-02-03-todo-api-design.md).

### Key Patterns Demonstrated

- **Aggregate Root** - Todo entity enforcing business invariants
- **Value Objects** - Type-safe domain concepts (TaskTitle, Priority, DueDate)
- **Domain Events** - Decoupled communication between aggregates
- **Repository Pattern** - Abstracted data access via ports
- **Transactional Outbox** - Reliable event publishing
- **Dependency Injection** - Manual wiring in `cmd/todo/main.go`
- **Testing Strategy** - Unit tests, integration tests with testcontainers, and API tests

## Go Directories

### `/cmd`

Main applications for this project.

The directory name for each application should match the name of the executable you want to have (e.g., `/cmd/myapp`).

Don't put a lot of code in the application directory. If you think the code can be imported and used in other projects, then it should live in the `/pkg` directory. If the code is not reusable or if you don't want others to reuse it, put that code in the `/internal` directory. You'll be surprised what others will do, so be explicit about your intentions!

It's common to have a small `main` function that imports and invokes the code from the `/internal` and `/pkg` directories and nothing else.

See the [`/cmd`](cmd/README.md) directory for examples.

### `/internal`

Private application and library code. This is the code you don't want others importing in their applications or libraries. Note that this layout pattern is enforced by the Go compiler itself. See the Go 1.4 [`release notes`](https://golang.org/doc/go1.4#internalpackages) for more details. Note that you are not limited to the top level `internal` directory. You can have more than one `internal` directory at any level of your project tree.

You can optionally add a bit of extra structure to your internal packages to separate your shared and non-shared internal code. It's not required (especially for smaller projects), but it's nice to have visual clues showing the intended package use. Your actual application code can go in the `/internal/app` directory (e.g., `/internal/app/myapp`) and the code shared by those apps in the `/internal/pkg` directory (e.g., `/internal/pkg/myprivlib`).

You use internal directories to make packages private. If you put a package inside an internal directory, then other packages can’t import it unless they share a common ancestor. And it’s the only directory named in Go’s documentation and has special compiler treatment.

### `/pkg`

Library code that's ok to use by external applications (e.g., `/pkg/mypubliclib`). Other projects will import these libraries expecting them to work, so think twice before you put something here :-) Note that the `internal` directory is a better way to ensure your private packages are not importable because it's enforced by Go. The `/pkg` directory is still a good way to explicitly communicate that the code in that directory is safe for use by others. The [`I'll take pkg over internal`](https://travisjeffery.com/ill-take-pkg-over-internal/) blog post by Travis Jeffery provides a good overview of the `pkg` and `internal` directories and when it might make sense to use them.

It's also a way to group Go code in one place when your root directory contains lots of non-Go components and directories making it easier to run various Go tools (as mentioned in these talks: [`Best Practices for Industrial Programming`](https://www.youtube.com/watch?v=PTE4VJIdHPg) from GopherCon EU 2018, [GopherCon 2018: Kat Zien - How Do You Structure Your Go Apps](https://www.youtube.com/watch?v=oL6JBUk6tj0) and [GoLab 2018 - Massimiliano Pippi - Project layout patterns in Go](https://www.youtube.com/watch?v=3gQa1LWwuzk)).

See the [`/pkg`](pkg/README.md) directory if you want to see which popular Go repos use this project layout pattern. This is a common layout pattern, but it's not universally accepted and some in the Go community don't recommend it.

It's ok not to use it if your app project is really small and where an extra level of nesting doesn't add much value (unless you really want to :-)). Think about it when it's getting big enough and your root directory gets pretty busy (especially if you have a lot of non-Go app components).

The `pkg` directory origins: The old Go source code used to use `pkg` for its packages and then various Go projects in the community started copying the pattern (see [`this`](https://twitter.com/bradfitz/status/1039512487538970624) Brad Fitzpatrick's tweet for more context).

### `/vendor`

Application dependencies (managed manually or by your favorite dependency management tool like the new built-in [`Go Modules`](https://go.dev/wiki/Modules) feature). The `go mod vendor` command will create the `/vendor` directory for you. Note that you might need to add the `-mod=vendor` flag to your `go build` command if you are not using Go 1.14 where it's on by default.

Don't commit your application dependencies if you are building a library.

Note that since [`1.13`](https://golang.org/doc/go1.13#modules) Go also enabled the module proxy feature (using [`https://proxy.golang.org`](https://proxy.golang.org) as their module proxy server by default). Read more about it [`here`](https://blog.golang.org/module-mirror-launch) to see if it fits all of your requirements and constraints. If it does, then you won't need the `vendor` directory at all.

## Service Application Directories

### `/api`

OpenAPI/Swagger specs, JSON schema files, protocol definition files.

See the [`/api`](api/README.md) directory for examples.

## Web Application Directories

### `/web`

Web application specific components: static web assets, server side templates and SPAs.

## Common Application Directories

### `/configs`

Configuration file templates or default configs.

Put your `confd` or `consul-template` template files here.

### `/init`

System init (systemd, upstart, sysv) and process manager/supervisor (runit, supervisord) configs.

### `/scripts`

Scripts to perform various build, install, analysis, etc operations.

These scripts keep the root level Makefile small and simple (e.g., [`https://github.com/hashicorp/terraform/blob/main/Makefile`](https://github.com/hashicorp/terraform/blob/main/Makefile)).

See the [`/scripts`](scripts/README.md) directory for examples.

### `/build`

Packaging and Continuous Integration.

Put your cloud (AMI), container (Docker), OS (deb, rpm, pkg) package configurations and scripts in the `/build/package` directory.

Put your CI (travis, circle, drone) configurations and scripts in the `/build/ci` directory. Note that some of the CI tools (e.g., Travis CI) are very picky about the location of their config files. Try putting the config files in the `/build/ci` directory linking them to the location where the CI tools expect them (when possible).

### `/deployments`

IaaS, PaaS, system and container orchestration deployment configurations and templates (docker-compose, kubernetes/helm, terraform). Note that in some repos (especially apps deployed with kubernetes) this directory is called `/deploy`.

### `/test`

Additional external test apps and test data. Feel free to structure the `/test` directory anyway you want. For bigger projects it makes sense to have a data subdirectory. For example, you can have `/test/data` or `/test/testdata` if you need Go to ignore what's in that directory. Note that Go will also ignore directories or files that begin with "." or "_", so you have more flexibility in terms of how you name your test data directory.

See the [`/test`](test/README.md) directory for examples.

## Other Directories

### `/docs`

Design and user documents (in addition to your godoc generated documentation).

See the [`/docs`](docs/README.md) directory for examples.

### `/tools`

Supporting tools for this project. Note that these tools can import code from the `/pkg` and `/internal` directories.

See the [`/tools`](tools/README.md) directory for examples.

### `/examples`

Examples for your applications and/or public libraries.

See the [`/examples`](examples/README.md) directory for examples.

### `/third_party`

External helper tools, forked code and other 3rd party utilities (e.g., Swagger UI).

### `/githooks`

Git hooks.

### `/assets`

Other assets to go along with your repository (images, logos, etc).

### `/website`

This is the place to put your project's website data if you are not using GitHub pages.

See the [`/website`](website/README.md) directory for examples.

## Directories You Shouldn't Have

### `/src`

Some Go projects do have a `src` folder, but it usually happens when the devs came from the Java world where it's a common pattern. If you can help yourself try not to adopt this Java pattern. You really don't want your Go code or Go projects to look like Java :-)

Don't confuse the project level `/src` directory with the `/src` directory Go uses for its workspaces as described in [`How to Write Go Code`](https://golang.org/doc/code.html). The `$GOPATH` environment variable points to your (current) workspace (by default it points to `$HOME/go` on non-windows systems). This workspace includes the top level `/pkg`, `/bin` and `/src` directories. Your actual project ends up being a sub-directory under `/src`, so if you have the `/src` directory in your project the project path will look like this: `/some/path/to/workspace/src/your_project/src/your_code.go`. Note that with Go 1.11 it's possible to have your project outside of your `GOPATH`, but it still doesn't mean it's a good idea to use this layout pattern.
