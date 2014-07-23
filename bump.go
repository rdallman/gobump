package main

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
	VERSION = "hi"
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

									s, err := fi.Stat()
									fmt.Println(s.Size())

									r := bufio.NewReader(fi)
									n, err := io.CopyN(ioutil.Discard, r, int64(offset+len)) // scrap the old
									n, err = fi.Seek(int64(offset), 0)                       // less error prone
									nt, err := io.WriteString(fi, `"hi"`)                    // write the new
									n, err = io.Copy(fi, r)                                  // write the rest
									_, _, _ = n, nt, err
									//fmt.Println(w.Buffered())
									//fmt.Println(n, err)
									//err = fi.Truncate(0)
									//fmt.Println(w.Buffered())
									//fmt.Println(err) // print err, but disregard
									//err = w.Flush()
									//fmt.Println(w.Buffered())
									//fmt.Println(err)
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
