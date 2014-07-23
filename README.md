# gobump

Inspired by https://rubygems.org/gems/bump

### Install

`go install github.com/rdallman/gobump`

### Quick Usage

In any file of a go package, at the top level:

```go
const VERSION = "0.0.1"
```

When you need to bump the version:

```shell
$ gobump
```

### Motivation

Go binaries shouldn't rely on `meta.json` files, why not bake the version into the
package with a usable variable from the start?

### Detailed VERSION info

```
VERSION is of the form:

    XX.YY.ZZ

 Where  XX == Major Version
        YY == Minor Version
        ZZ == Patch Version

 XX is mandatory, whereas YY and ZZ are optional.
 Additionally, there can be no ZZ without a YY.
 If invoking the bumping of a patch or minor version
 when one does not exist, all necessary fields will be created
 and the rightmost incremented. Accordingly, all values
 to the right of the one being incremented will be cleared.

 XX, YY, and ZZ can have unbounded length
 within the limitations of the given machine's architecture.

 Invoking gobump on a package will only increment XX, YY or ZZ by one.
 They are assumed to be numeric and puppies will die if you try to version
 otherwise.

 VERSION can only be interpreted as a CONST at the package level,
 and its declaration may be placed in any file within a package.

 Below is a valid VERSION visible and modifiable by the gobump program.
 So this program can accept itself as input. Big woop.
```
