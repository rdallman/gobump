package main

//

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	VERSION = "3.0.10"
)

// TODO look at gopkg.in versioning

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("strange things are afoot at the circle K")
	}

	f := token.NewFileSet()
	pkgs, err := parser.ParseDir(f, pwd, nil, 0)

	var newv string

OG:
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
									// bump it

									vf := f.Position(spec.Values[i].Pos())
									fname := vf.Filename
									offset := vf.Offset
									len := f.Position(spec.Values[i].End()).Offset - offset

									fmt.Println(fname, offset, len)

									fi, err := os.OpenFile(fname, os.O_RDWR, 0666)
									if err != nil {
										log.Fatalln("don't go in there ted")
									}
									defer fi.Close()

									r := bufio.NewReader(fi)
									w := bufio.NewWriter(fi)
									n, err := io.CopyN(w, r, int64(offset)) // copy up to version
									fmt.Println(n, err)
									n, err = io.CopyN(ioutil.Discard, r, int64(len)) // scrap the old
									fmt.Println(n, err)
									nt, err := io.WriteString(w, "hi") // write the new
									fmt.Println(nt, err)
									n, err = io.Copy(w, r) // write the rest
									fmt.Println(n, err)
									err = fi.Truncate(0)
									fmt.Println(err) // print err, but disregard
									err = w.Flush()
									fmt.Println(err)
									break OG
								}
							}
						}
					}
				}
			}
		}
	}
	fmt.Println(newv)
}
