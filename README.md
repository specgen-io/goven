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
go install github.com/specgen-io/goven@v0.0.2
```

## Usage

To vendor only `replace` directives in the module, change name of the module to `github.com/example/mymodule` run this command in the module root folder: 

```
goven -name github.com/example/mymodule
```

The module code with vendored dependencies will be placed into `out` subfolder. The code in `out` subfolder will be ready for publishing at `github.com/example/mymodule`.

The output folder could be changed by `-out` option:

```
goven -name github.com/example/mymodule -out ~/mymodule
```

To vendor both `replace` and `require` directives, run following:

```
go mod vendor    # <- this is still needed
goven -name github.com/example/mymodule -required
```

Vendored modules are by default placed to `goven` subfolder of the module. This can be change via `-vendor` option.

```
goven -name github.com/example/mymodule -required -vendor vendored
```

Check `goven -help` to learn about all arguments.
