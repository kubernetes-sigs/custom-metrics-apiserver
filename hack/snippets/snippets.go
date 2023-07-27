// Copyright 2022 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
A simple tool that prints snippets from Go code.

Usage:

	go run ./hack/snippets -d [DECLARATIONS_LIST] ./path/to/file.go

DECLARATIONS_LIST is a comma-separated list of top-level declarations
identifiers:

  - package: prints the package declaration (with package comment)

  - import: prints the FIRST import block

  - func=[receiver.]name prints the function or method named "name".

    For functions, simply use the name, e.g. func=myFunc

    For methods, specify also the receiver type, e.g. func=myType.myFunc or func=*myType.myFunc

  - var=name prints the var block where var "name" is declared

  - cons=name prints the const block where const "name" is declared

  - type=name prints the type block where type "name" is declared

Top-level declarations are printed, along with their comments, in the order
specified with the "d" flag.
*/
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

type decl struct {
	kind string
	name string
	rcv  string
}

func parseArg(d string) ([]decl, error) {
	if d == "" {
		return []decl{}, errors.New("empty")
	}
	l := strings.Split(d, ",")
	res := make([]decl, len(l))
	for i, s := range l {
		kind, name, _ := strings.Cut(s, "=")
		rcv := ""

		switch kind {
		case "func":
			if name == "" {
				return []decl{}, errors.New("func without func name")
			}
			rcvSegment, nameSegment, hasRcv := strings.Cut(name, ".")
			if hasRcv {
				name = nameSegment
				rcv = rcvSegment
			}
		case "var", "const", "type":
			if name == "" {
				return []decl{}, fmt.Errorf("%s without %s name", kind, kind)
			}
		case "package", "import":
		default:
			return []decl{}, fmt.Errorf("invalid declaration type %q", kind)
		}

		res[i] = decl{kind: kind, name: name, rcv: rcv}
	}
	return res, nil
}

func matchesType(t ast.Expr, name string) bool {
	if strings.HasPrefix(name, "*") {
		if starExpr, ok := t.(*ast.StarExpr); ok {
			return matchesType(starExpr.X, strings.TrimPrefix(name, "*"))
		}
	} else {
		if i, ok := t.(*ast.Ident); ok {
			return i.Name == name
		}
	}
	return false
}

func matchesFuncDecl(decl ast.Decl, d decl) (*ast.FuncDecl, bool) {
	if d.kind != "func" {
		return nil, false
	}

	if fun, ok := decl.(*ast.FuncDecl); ok {
		if fun.Name.String() == d.name {
			if d.rcv != "" && fun.Recv != nil {
				for _, r := range fun.Recv.List {
					if matchesType(r.Type, d.rcv) {
						return fun, true
					}
				}
			} else if d.rcv == "" && fun.Recv == nil {
				return fun, true
			}
		}
	}
	return nil, false
}

func matchesGenDecl(decl ast.Decl, d decl) (*ast.GenDecl, bool) {
	if gen, ok := decl.(*ast.GenDecl); ok {
		if d.kind == "type" && gen.Tok == token.TYPE {
			for _, s := range gen.Specs {
				typeSpec, ok := s.(*ast.TypeSpec)
				if ok && typeSpec.Name.String() == d.name {
					return gen, true
				}
			}
		}
		if (d.kind == "var" && gen.Tok == token.VAR) || (d.kind == "const" && gen.Tok == token.CONST) {
			for _, s := range gen.Specs {
				valueSpec, ok := s.(*ast.ValueSpec)
				if ok {
					for _, name := range valueSpec.Names {
						if name.String() == d.name {
							return gen, true
						}
					}
				}
			}
		}
		if d.kind == "import" && gen.Tok == token.IMPORT {
			return gen, true
		}
	}
	return nil, false
}

func find(fset *token.FileSet, f *ast.File, d decl) (start token.Position, end token.Position, found bool) {
	if d.kind == "package" {
		found = true
		doc := f.Doc

		pos := f.Name.Pos()
		if doc != nil && doc.Pos() < pos {
			pos = doc.Pos()
		}
		start = fset.PositionFor(pos, false)

		endPos := f.Name.End()
		end = fset.PositionFor(endPos, false)
		return
	}

	for _, decl := range f.Decls {
		var doc *ast.CommentGroup

		// Either a FuncDecl or GenDecl

		if fun, ok := matchesFuncDecl(decl, d); ok {
			doc = fun.Doc
			found = true
		}
		if gen, ok := matchesGenDecl(decl, d); ok {
			doc = gen.Doc
			found = true
		}

		if found {
			pos := decl.Pos()
			if doc != nil && doc.Pos() < pos {
				pos = doc.Pos()
			}
			start = fset.PositionFor(pos, false)

			endPos := decl.End()
			end = fset.PositionFor(endPos, false)

			return
		}
	}

	return token.Position{}, token.Position{}, false
}

func printSnippets(w io.Writer, r io.Reader, filename string, decls []decl) error {
	inputBuf := bytes.Buffer{}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, io.TeeReader(r, &inputBuf), parser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot parse Go file %s: %w", filename, err)
	}

	b := inputBuf.Bytes()
	outputBuf := &bytes.Buffer{}

	for i, decl := range decls {
		start, end, found := find(fset, f, decl)
		if !found {
			return fmt.Errorf("declaration not found: %v", decl)
		}

		sc := bufio.NewScanner(bytes.NewBuffer(b))
		var line int
		for sc.Scan() {
			line++
			if line >= start.Line && line <= end.Line {
				fmt.Fprintln(outputBuf, sc.Text())
			}
		}

		// Add an empty line between two declarations
		if i < len(decls)-1 {
			fmt.Fprintln(outputBuf)
		}
	}

	_, err = io.Copy(w, outputBuf)
	return err
}

func run(filename string, d string) error {
	if filename == "" {
		return errors.New("no input file")
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", filename, err)
	}
	defer file.Close()

	decls, err := parseArg(d)
	if err != nil {
		return fmt.Errorf("invalid flag d=%q: %w", d, err)
	}

	return printSnippets(os.Stdout, file, filename, decls)
}

func main() {
	var d string
	flag.StringVar(&d, "d", "", "declaration")
	flag.Parse()

	filename := flag.Arg(0)

	if err := run(filename, d); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
