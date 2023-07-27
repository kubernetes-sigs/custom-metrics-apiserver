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

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArg(t *testing.T) {
	cases := map[string]struct {
		arg         string
		expectedErr bool
		exptected   []decl
	}{
		// errors
		"empty":           {"", true, []decl{}},
		"funcWithoutName": {"func", true, []decl{}},
		"varWithoutName":  {"var", true, []decl{}},
		"invalid":         {"foo", true, []decl{}},

		// success
		"singleFunc": {"func=foo", false, []decl{{kind: "func", name: "foo"}}},
		"methods": {"func=my.Add,func=*my.Close", false, []decl{
			{kind: "func", name: "Add", rcv: "my"},
			{kind: "func", name: "Close", rcv: "*my"},
		}},
		"multiple": {"import,func=foo", false, []decl{
			{kind: "import"},
			{kind: "func", name: "foo"},
		}},
	}
	for caseName, c := range cases {
		t.Run(caseName, func(t *testing.T) {
			res, err := parseArg(c.arg)
			if c.expectedErr {
				assert.Error(t, err)
			} else if assert.NoError(t, err) {
				assert.Equal(t, c.exptected, res)
			}
		})
	}
}

const file = `
// A top level comment

// package foo is a sample package
package foo

import (
	"fmt"
)

// Operation represents a mathematical operation.
//
// It has two functions, to compute then print the result.
type Operation interface {
	Compute()
	PrintResult()
}

type (
	mySum struct {
		a      myInt
		b      myInt
		result myInt
	}

	// myInt is a type alias.
	myInt int
)

// sum calcutates a sum.
func sum(a myInt, b myInt) myInt {
	return a + b
}

func (m *mySum) Compute() {
	m.result = sum(m.a, m.b)
}

func (m mySum) PrintResult() {
	fmt.Printf("%v + %v = %v", m.a, m.b, m.result)
}

// n is a constant
const n = 1

var (
	_    Operation = &mySum{}
	sum1           = &mySum{a: n, b: n}
)

func Compute() {
	sum1.Compute()
}

func PrintResult() {
	sum1.PrintResult()
}
`

func TestSnippets(t *testing.T) {
	cases := map[string]struct {
		decls       []decl
		expectedErr bool
		exptected   string
	}{
		"notFound": {
			decls:       []decl{{kind: "func", name: "foo"}},
			expectedErr: true, // function does not exist
		},
		"notFoundReceiver": {
			decls:       []decl{{kind: "func", name: "Compute", rcv: "mySum"}},
			expectedErr: true, // methods *mySum.Compute has not the same receiver
		},
		"threeFuncs": {
			decls: []decl{
				{kind: "func", name: "PrintResult"},
				{kind: "func", name: "Compute"},
				{kind: "func", name: "Compute", rcv: "*mySum"},
			},
			exptected: `func PrintResult() {
	sum1.PrintResult()
}

func Compute() {
	sum1.Compute()
}

func (m *mySum) Compute() {
	m.result = sum(m.a, m.b)
}
`,
		},
		"funcWithComment": {
			decls: []decl{{kind: "func", name: "sum"}},
			exptected: `// sum calcutates a sum.
func sum(a myInt, b myInt) myInt {
	return a + b
}
`,
		},
		"type": {
			decls: []decl{{kind: "type", name: "Operation"}},
			exptected: `// Operation represents a mathematical operation.
//
// It has two functions, to compute then print the result.
type Operation interface {
	Compute()
	PrintResult()
}
`,
		},
		"varInVarBlock": {
			decls: []decl{{kind: "var", name: "sum1"}},
			exptected: `var (
	_    Operation = &mySum{}
	sum1           = &mySum{a: n, b: n}
)
`,
		},
		"constWithComment": {
			decls: []decl{{kind: "const", name: "n"}},
			exptected: `// n is a constant
const n = 1
`,
		},
		"packandAndImports": {
			decls: []decl{{kind: "package"}, {kind: "import"}},
			exptected: `// package foo is a sample package
package foo

import (
	"fmt"
)
`,
		},
	}

	for caseName, c := range cases {
		t.Run(caseName, func(t *testing.T) {
			var b bytes.Buffer
			err := printSnippets(&b, strings.NewReader(file), "file.go", c.decls)
			if c.expectedErr {
				assert.Error(t, err)
			} else if assert.NoError(t, err) {
				assert.Equal(t, c.exptected, b.String())
			}
		})
	}
}
