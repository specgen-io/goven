# goven

The golang module vendoring tool.

## Why?

Go has vendoring functionality built-in, so why we need goven?

Yes, go tooling has vendoring functionality. However some go tools do not respect vendored modules. For example, `go install` doesn't take into account vendored modules in the module being installed.

Also some of your modules might have `replace` directives in `go.mod` file. This would make your module non usable for `go get` or `go install` commands.

If simple go vendoring is working for you, please keep using it.

## How?

How goven is vendoring modules?

Goven copies all dependencies source code into the code of the module, it also fixes all package references in the code and fixing `go.mod` file.

To use goven for vendoring, the module should be developed in one repository but released via the other repository.

## Installation
```
go install github.com/specgen-io/goven@v0.0.9
```

## Usage

There are two commands implemented by goven: `vendor` and `release`.

The `vendor` command vendors the module code and places it to the local file system.

While the `release` command does the same as `vendor` command but also pushes vendored code to the Github.

### Vendor

To vendor only `replace` directives in the module, change name of the module to `github.com/example/mymodule` run this command in the module root folder: 

```
goven vendor -name github.com/example/mymodule
```

The module code with vendored dependencies will be placed into `out` subfolder.
The code in `out` subfolder will be ready for publishing at `github.com/example/mymodule`.

The output folder could be changed by `-out` option:

```
goven vendor -name github.com/example/mymodule -out ~/mymodule
```

To vendor both `replace` and `require` directives, run following:

```
go mod vendor    # <- this is still needed
goven vendor -name github.com/example/mymodule -required
```

Vendored modules are by default placed to `goven` subfolder of the module.
This can be change via `-vendor` option:

```
goven vendor -name github.com/example/mymodule -required -vendor vendored
```

Check `goven vendor -help` to learn about all arguments.

### Release

This command will vendor code rename the module into `github.com/example/mymodule` and push it to the corresponding Github repository:

```
goven release -name github.com/example/mymodule
```

In the case below the vendored module will be named `github.com/example/mymodule/v2` and also it will be pushed into `v2` folder of the repository as go requires:

```
goven release -name github.com/example/mymodule/v2
```

The `release` command can create and push tag into the release repository.
This could be done via `-version` option:

```
goven release -name github.com/example/mymodule/v2 -version v2.0.0
```

The other options like `-module`, `-required`, `-out`, `-vendor` are working for `release` command the same way as for `vendor` command.

Committing and pushing to Github repository requires authorization, options `github-name`, `github-email`, `github-user` and `github-token` should be used to provide credentials and commit author's info.
Alternatively these options could be set via `GITHUB_NAME`, `GITHUB_EMAIL`, `GITHUB_USER` and `GITHUB_TOKEN` environment variables.

Check `goven release -help` to learn about all arguments.
