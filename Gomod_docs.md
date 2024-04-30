# Go.mod Documentation

## Overview

The **go.mod** file is an essential part of any Go project. It defines the module’s dependencies, version constraints, and other configuration settings. Let’s dive into the details: 

---
## File Structure

The **go.mod** file typically resides at the root of your project directory. Its primary purpose is to manage the module’s dependencies and ensure reproducible builds.

## Syntax
The go.mod file follows a straightforward syntax:
``` Go
    module <module-name>

    go <go-version>
```
### Example
```Go
    module icapeg

    go 1.19
```

To update the version of Go to go 1.20 run this code:

```Go
go mod edit -go 1.20
```

you can choose any version you want.

## Dependencies
Dependencies are listed in the following format:

``` Go
    require (
        <module-path> <version>
        // Additional dependencies...
    )
```

### Example

```Go
    require (
        github.com/davecgh/go-spew v1.1.1
        github.com/dutchcoders/go-clamd v0.0.0-20170520113014-b970184f4d9e
        github.com/h2non/filetype v1.0.12
        github.com/spf13/viper v1.9.0
        github.com/xhit/go-str2duration/v2 v2.0.0
        go.uber.org/zap v1.22.0
    )
```

To update dependency run this code:

```Go
go get <module-path> @ <version>
```

### Example :

```Go
go get github.com/spf13/viper@latest
```

To specify the version we want to use simply use this code:
```Go
go get github.com/spf13/viper@v1.9.0
```
## Update All Modules to Latest Patch Version:
If you want to update all Go modules to the latest patch version (i.e., only minor and patch version changes), use the following command:
```Go
go get -u=patch ./...
```
This command ensures that packages are upgraded only to the latest patch version according to Semantic Versioning (MAJOR.MINOR.PATCH) 

## Update All Modules Recursively:
To recursively update packages in all subdirectories, use:

```Go
go get -u ./...
```

## Tips
- Run `go mod tidy` to add missing and remove unused dependencies.
- Use `go list -m all` to view all dependencies and their versions.




