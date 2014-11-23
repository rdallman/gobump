// package gobump is a CLI tool for bumping semantic versions of go packages.
//
// gobump expects semantic versioning, where version
// is a package level constant named `Version` of the form:
//
//    const Version = 'XX.YY.ZZ'
//
// Where:
//
//        XX == Major Version
//        YY == Minor Version
//        ZZ == Patch Version
//
// XX is mandatory, whereas YY and ZZ are optional.
// Additionally, there can be no ZZ without a YY.
// If invoking the bumping of a patch or minor version
// when one does not exist, all necessary fields will be created
// and the rightmost incremented. Accordingly, all values
// to the right of the one being incremented will be cleared.
//
// XX, YY, and ZZ can have unbounded length.
// ...within the limitations of the given machine's architecture.
//
// Invoking `gobump` on a package will only increment XX, YY or ZZ by one.
// They are assumed to be numeric and puppies will die if you try to version
// otherwise.
//
// `Version` can only be interpreted as a const at the package level,
// and its declaration may be placed in any file within a package.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rdallman/gobump/bump"
)

// TODO detect git|mercurial|bzr|none
// TODO hah, const must have value. doh. remove all that logic
// TODO look at gopkg.in versioning, consider compatibility

const (
	// Below is a valid `Version`, visible and modifiable by the gobump program.
	// So this program can accept itself as input. Big woop.
	Version = "0.0.16"
)

var (
	nocommit = flag.Bool("no-commit", false, "don't make a commit after bumping, commit by default")
	notag    = flag.Bool("no-tag", false, "don't tag the new commit, by default tags the current or new commit")
	help     = flag.Bool("help", false, "show this message")
	help2    = flag.Bool("h", false, "show this message")
)

const cmd = `usage: gobump [flags] [major|minor|patch] [go package]`

func usage() {
	fmt.Fprintln(os.Stderr, cmd, `

If [major|minor|patch] is not given, defaults to "patch"

If [go package] is not given, defaults to $PWD
`)

	flag.PrintDefaults()
	os.Exit(1)
}

func shortUsage() {
	fmt.Fprintln(os.Stderr, cmd)
	os.Exit(1)
}

func main() {
	flag.Parse()
	args := flag.Args()

	if *help || *help2 {
		usage()
	}

	h := bump.Patch
	if len(args) > 0 {
		switch args[0] {
		case "major":
			h = bump.Major
		case "minor":
			h = bump.Minor
		case "patch":
		default: // invalid input,
			shortUsage()
		}
	}

	var pkg string
	if len(args) < 2 {
		var err error
		pkg, err = os.Getwd()
		if err != nil {
			log.Fatalln("strange things are afoot at the circle K:", err)
		}
	} else {
		pkg = args[1]
	}

	fname, bumped, err := bump.Bump(h, pkg)
	if err != nil {
		log.Fatalln(err)
	}

	var out bool
	if !*nocommit {
		out = true
		bump.GitCommit(fname, bumped)
	}
	if !*notag {
		bump.GitTag(bumped)
	}

	if !out {
		fmt.Println(bumped)
	}
}
