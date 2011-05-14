package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
)

func main() {
	flag.Parse()
	filenames := flag.Args()
	fset := token.NewFileSet()
	pkgMap, firstErr := parser.ParseFiles(fset, filenames, 0)
	if firstErr != nil {
		fmt.Fprintf(os.Stderr, "Error while parsing: %v\n", firstErr)
	}

	v := NewNodeChecker(fset)
	v.InterfaceName = regexp.MustCompile("I[A-Z][A-Za-z]+")

	for _, pkg := range pkgMap {
		ast.Walk(v, pkg)
	}
}

type NodeChecker struct {
	fset          *token.FileSet
	InterfaceName *regexp.Regexp
}

func NewNodeChecker(fset *token.FileSet) *NodeChecker {
	return &NodeChecker{
		fset: fset,
	}
}

func (v *NodeChecker) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.TypeSpec:
		v.checkTypeName(n)
	}
	return v
}

// report displays a message about a particular position in the fileset.
func (v *NodeChecker) report(pos token.Pos, format string, args ...interface{}) {
	position := v.fset.Position(pos)
	allArgs := make([]interface{}, len(args)+3)

	allArgs[0] = position.Filename
	allArgs[1] = position.Line
	allArgs[2] = position.Column
	copy(allArgs[3:], args)

	fmt.Fprintf(
		os.Stderr,
		"%s:%d:%d: "+format,
		allArgs...)
}

func (v *NodeChecker) checkTypeName(typeSpec *ast.TypeSpec) {
	name := typeSpec.Name.Name
	switch t := typeSpec.Type.(type) {
	case *ast.InterfaceType:
		if !v.InterfaceName.MatchString(name) {
			v.report(typeSpec.Name.NamePos, "Bad name for interface %q\n", name)
		}
	}
}
