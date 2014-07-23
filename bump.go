// Package provides simple version bumping CLI that is kept inside a go variable for use.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// TODO hah, const must have value. doh. remove all that logic
// TODO look at gopkg.in versioning, consider compatibility
// TODO usage() info in --help

const (
	// VERSION is of the form:
	//
	//    XX.YY.ZZ
	//
	// Where  XX == Major Version
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
	// XX, YY, and ZZ can have unbounded length
	// within the limitations of the given machine's architecture.
	//
	// Invoking gobump on a package will only increment XX, YY or ZZ by one.
	// They are assumed to be numeric and puppies will die if you try to version
	// otherwise.
	//
	// VERSION can only be interpreted as a CONST at the package level,
	// and its declaration may be placed in any file within a package.
	//
	// Below is a valid VERSION visible and modifiable by the gobump program.
	// So this program can accept itself as input. Big woop.
	//
	VERSION = "0.0.2"
)

type howhigh byte

const (
	Major howhigh = iota
	Minor
	Patch
)

var (
	commit = flag.Bool("commit", false, "make a commit after bumping")
	tag    = flag.Bool("tag", false, "tag commit, if combined with --commit will tag that commit")
)

func main() {
	flag.Parse()
	args := flag.Args()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("strange things are afoot at the circle K")
	}

	h := Patch
	if len(args) > 0 {
		switch args[0] {
		case "major":
			h = Major
		case "minor":
			h = Minor
		case "patch":
			h = Patch
		default:
			log.Fatalln("invalid bump, can only be one of [major, minor, patch]")
		}
	}

	fname, bump, err := Bump(h, pwd)
	if err != nil {
		log.Fatalln(err)
	}

	var out bool
	// TODO git commit, tag
	if *commit {
		out = true
		gitcommit(fname, bump)
	}
	if *tag {
		gittag(bump)
	}

	if !out {
		fmt.Println(bump)
	}
}

func gitcommit(fname, version string) {
	out, _ := exec.Command("git", "add", fname).CombinedOutput()
	fmt.Printf("%s", string(out))
	out, _ = exec.Command("git", "commit", "-m", version).CombinedOutput()
	fmt.Printf("%s", string(out))
}

func gittag(version string) {
	out, _ := exec.Command("git", "tag", version).CombinedOutput()
	fmt.Printf("%s", string(out))
}

func Bump(h howhigh, pkg string) (fname, version string, err error) {
	fset, pos, end, err := findVersion(pkg)
	if err != nil {
		return "", "", err
	}

	fname, offset, len := extractInfos(fset, pos, end)

	f, err := os.OpenFile(fname, os.O_RDWR, 0666)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	delimb, delime, old, err := extractOld(f, offset, len)
	if err != nil {
		return "", "", err
	}

	newv, err := bump(h, old)
	if err != nil {
		return "", "", err
	}
	fullversion := delimb + newv + delime

	err = writeNew(f, fullversion, offset, len)
	if err != nil {
		return "", "", err
	}

	return fname, newv, nil
}

func writeNew(f *os.File, v string, offset, length int64) error {
	_, err := f.Seek(offset, 0) // less error prone
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	// TODO this is just off, need to get a separate reader and writer for file
	r := io.NewSectionReader(f, offset+int64(len(v)), stat.Size()-offset+length)
	_, err = f.Seek(offset, 0) // less error prone
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, v) // write the new
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r) // write the rest
	if err != nil {
		return err
	}

	//f.Sync()
	return nil
}

// Increments "XX.YY.ZZ" appropriately, expected input has string delimiters removed.
// Empty string is valid.
func bump(h howhigh, old string) (newv string, err error) {
	var s []string
	if old != "" {
		s = strings.Split(old, ".")
	}

	vs := make([]int, len(s))
	for i, m := range s {
		vs[i], err = strconv.Atoi(m)
		if err != nil {
			return "", err
		}
	}

	var news []string

	switch h {
	case Major:
		switch len(s) {
		case 0:
			news = []string{"1"}
		case 1:
			news = []string{strconv.Itoa(vs[0] + 1)}
		case 2:
			news = []string{strconv.Itoa(vs[0] + 1), "0"}
		case 3:
			news = []string{strconv.Itoa(vs[0] + 1), "0", "0"}
		}
	case Minor:
		switch len(s) {
		case 0:
			news = []string{"0", "1"}
		case 1:
			news = []string{s[0], "1"}
		case 2:
			news = []string{s[0], strconv.Itoa(vs[1] + 1)}
		case 3:
			news = []string{s[0], strconv.Itoa(vs[1] + 1), "0"}
		}
	case Patch:
		switch len(s) {
		case 0:
			news = []string{"0", "0", "1"}
		case 1:
			news = []string{s[0], "0", "1"}
		case 2:
			news = []string{s[0], s[1], "1"}
		case 3:
			news = []string{s[0], s[1], strconv.Itoa(vs[2] + 1)}
		}
	}

	if news == nil {
		return "", errors.New("way too high")
	}

	return strings.Join(news, "."), nil
}

// extractInfos returns all info we need from ast at source level.
func extractInfos(f *token.FileSet, pos, end token.Pos) (fname string, offset, len int64) {
	beg := f.Position(pos)
	fname = beg.Filename
	offset = int64(beg.Offset)
	len = int64(f.Position(end).Offset) - offset

	return fname, offset, len
}

// extractOld returns non-delimited old version with the accompanying delimiters, if any.
// If value was uninitialized, we'll add an equals sign and delimiters.
func extractOld(fi *os.File, offset, length int64) (delimb, delime, old string, err error) {
	if length == 0 {
		return ` = "`, `"`, "", nil
	}

	v := make([]byte, length)
	_, err = fi.ReadAt(v, offset)
	if err != nil {
		return "", "", "", err
	}
	delimb = string(v[0])
	delime = string(v[len(v)-1])
	return delimb, delime, string(v[1 : len(v)-1]), nil
}

// pos, end are position of beginning and end of value. if value is
// uninitialized, token.Pos of end of VERSION is returned in both places.
func findVersion(pkg string) (f *token.FileSet, pos, end token.Pos, err error) {
	// TODO consider benchmarking other approaches to parsing.
	f = token.NewFileSet()
	pkgs, err := parser.ParseDir(f, pkg, nil, 0)
	if err != nil {
		return nil, pos, end, err
	}

	// TODO see how much the scheduler would kill concurrizing this

	for _, pkg := range pkgs {
		pkgf := ast.MergePackageFiles(pkg, 0) // TODO could exclude as much as possible
		for _, d := range pkgf.Decls {
			switch d := d.(type) {
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					for _, spec := range d.Specs {
						switch spec := spec.(type) {
						case *ast.ValueSpec:
							for i, n := range spec.Names {
								if n.Name == "VERSION" {
									if spec.Values != nil {
										expr := spec.Values[i]
										return f, expr.Pos(), expr.End(), nil // we need to go deeper
									}
									return f, n.End(), n.End(), nil // return end of names, `= value` gets added
								}
							}
						}
					}
				}
			}
		}
	}
	return nil, pos, end, errors.New("Didn't find VERSION in package " + pkg)
}
