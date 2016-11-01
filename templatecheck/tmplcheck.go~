package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"os"
	"sync"
	tparse "text/template/parse"

	"golang.org/x/tools/go/loader"
)

var (
	TemplatesPath string
	PackagePath   string
	LeftDelim     string
	RightDelim    string
)

func main() {
	flag.StringVar(&TemplatesPath, "t", "", "path to templates directory")
	flag.StringVar(&PackagePath, "p", "", "package import path")
	flag.StringVar(&LeftDelim, "ldelim", "{{", "left delimiter in templates")
	flag.StringVar(&RightDelim, "rdelim", "}}", "right delimiter in templates")

	flag.Parse()

	if TemplatesPath == "" {
		fmt.Fprintln(os.Stderr, "-t is required")
		os.Exit(2)
	}
	if PackagePath == "" {
		fmt.Fprintln(os.Stderr, "-p is required")
		os.Exit(2)
	}

	var wg sync.WaitGroup

	var usages []usage
	var err0, err1 error

	wg.Add(1)
	go func() {
		defer wg.Done()
		usages, err0 = parsePackage(PackagePath)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err1 = parseTemplates(TemplatesPath)
	}()

	wg.Wait()

	fmt.Println(usages, err0)

	// b, err := ioutil.ReadFile("testdata/hello.html")
	// if err != nil {
	// fmt.Fprintln(os.Stderr, err)
	// os.Exit(-1)
	// }

	// m, err := tparse.Parse("hello", string(b), LeftDelim, RightDelim)

	// if err != nil {
	// fmt.Fprintln(os.Stderr, err)
	// os.Exit(-1)
	// }

	// fmt.Printf("%#v", m["hello"])

	// tree := m["hello"]
	// p(tree.Root)
}

// identValue returns the first value for the ident.
func identValue(a *ast.Ident) (string, error) {
	if a.Obj == nil || a.Obj.Decl == nil {
		return "", errors.New("failed to determine Decl")
	}
	vspec, ok := a.Obj.Decl.(*ast.ValueSpec)
	if !ok || len(vspec.Values) == 0 {
		return "", errors.New("unknown value")
	}
	return vspec.Values[0].(*ast.BasicLit).Value, nil
}

type call interface {
	Type() []string
	Func() []string
	Handler(callexpr *ast.CallExpr) (name string, keys []string, err error)
}

var errUnsupportedArgs = errors.New("unsupported type for arguments")

var templatePackages = []call{
	&templatesSet{},
	&htmltemplateTemplate{}, // Order matters: Support html/template before text/template.
	&texttemplateTemplate{},
}

type templatesSet struct{}

func (t *templatesSet) Type() []string {
	return []string{"github.com/go-web-framework/templates.Set", "*github.com/go-web-framework/templates.Set"}
}

func (t *templatesSet) Func() []string { return []string{"Execute"} }

func compositeLitKeys(comp *ast.CompositeLit) []string {
	var ret []string
	for _, e := range comp.Elts {
		k := e.(*ast.KeyValueExpr).Key
		switch x := k.(type) {
		case *ast.BasicLit:
			ret = append(ret, x.Value) // map
		case *ast.Ident:
			ret = append(ret, x.Name) // struct
		}
	}
	return ret
}

func identToCompositeLit(id *ast.Ident) (*ast.CompositeLit, error) {
	asn, ok := id.Obj.Decl.(*ast.AssignStmt)
	if !ok || len(asn.Rhs) == 0 {
		return nil, errors.New("identToCompositeLit: wrong type")
	}

	cl, ok := asn.Rhs[0].(*ast.CompositeLit)
	if !ok {
		return nil, errors.New("identToCompositeLit: wrong type")
	}

	return cl, nil
}

// Handler is the handler for templates.Set.
//
// Arguments support:
//
//   1. composite literals: Foo{X: Y}, map[KeyType]ValueType{x: y}
//   2. ident -> composite literal.
//
// If the composite literal is a map, only literal maps are supported. That is,
//
//   s.Execute(.., .., map[string]interface{} {
//     "qux": 2,
//     "bar": 10,
//   })
//
// is supported. But not:
//
//   m := map[string]interface{}
//   m["qux"] = 2
//   s.Execute(.., .., m)
//
func (t *templatesSet) Handler(callexpr *ast.CallExpr) (string, []string, error) {
	var name string
	var keys []string

	// Args[0] is the name of the template.

	switch a := callexpr.Args[0].(type) {
	case *ast.BasicLit:
		name = a.Value
	case *ast.Ident:
		n, err := identValue(a)
		if err != nil {
			return "", nil, err
		}
		name = n
	}

	// Args[2] is the arguments being passed.

	switch x := callexpr.Args[2].(type) {
	case *ast.CompositeLit:
		keys = compositeLitKeys(x)
	case *ast.Ident:
		c, err := identToCompositeLit(x)
		if err != nil {
			return "", nil, err
		}
		keys = compositeLitKeys(c)
	default:
		return "", nil, errUnsupportedArgs
	}

	return name, keys, nil
}

type htmltemplateTemplate struct{}

func (t *htmltemplateTemplate) Type() []string {
	return []string{"html/template.Template", "*html/template.Template"}
}

func (t *htmltemplateTemplate) Func() []string { return []string{"Execute"} }

func (t *htmltemplateTemplate) Handler(callexpr *ast.CallExpr) (string, []string, error) {
	return "", nil, errors.New("unsupported type")
}

type texttemplateTemplate struct{}

func (t *texttemplateTemplate) Type() []string {
	return []string{"text/template.Template", "*text/template.Template"}
}

func (t *texttemplateTemplate) Func() []string { return []string{"Execute"} }

func (t *texttemplateTemplate) Handler(callexpr *ast.CallExpr) (string, []string, error) {
	return "", nil, errors.New("unsupported type")
}

func match(typ, funcName string) (call, bool) {
	for _, tmpllib := range templatePackages {
		for _, t := range tmpllib.Type() {
			if t == typ {
				for _, f := range tmpllib.Func() {
					if f == funcName {
						return tmpllib, true
					}
				}
			}
		}
	}
	return nil, false
}

// usage represents a call to template with the keys.
type usage struct {
	Template string   // name of template
	Keys     []string // keys passed to template
}

// The return value is a slice of `usage` (not a map[Template]Args),
// because a single package can make multiple calls to the same template
// with different args.
func parsePackage(path string) ([]usage, error) {
	var conf loader.Config

	_, err := conf.FromArgs([]string{path}, false)
	if err != nil {
		return nil, err
	}

	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	ourpkg := prog.Package(path)

	var ret []usage
	var retErr error

	for _, f := range ourpkg.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				selexpr, ok := x.Fun.(*ast.SelectorExpr)
				if !ok {
					break
				}
				id, ok := selexpr.X.(*ast.Ident)
				if !ok {
					break
				}

				typ := ourpkg.TypeOf(id).String()
				funcName := selexpr.Sel.Name

				tl, ok := match(typ, funcName)
				if !ok {
					// Not a matching call. Move on to next call expression.
					break
				}

				name, keys, err := tl.Handler(x)
				if err != nil {
					retErr = err
					return false
				}
				ret = append(ret, usage{name, keys})
			}

			return true
		})
	}

	return ret, retErr
}

// TODO
func parseTemplates(path string) ([]usage, error) {
	return nil, nil
}

// TODO
func parseTemplate(fpath string) error {
	return nil
}

// TODO: DELETE
func p(listnode *tparse.ListNode) {
	for _, n := range listnode.Nodes {
		fmt.Println(n.Type())

		switch nn := n.(type) {
		case *tparse.ListNode:
			p(nn)
		case *tparse.CommandNode, *tparse.ChainNode, *tparse.FieldNode:
			fmt.Println("ccf")
			fmt.Println(nn)
		case *tparse.ActionNode:
			fmt.Println("act")
			fmt.Println(nn)
			fmt.Println(nn.Pipe.Cmds, nn.Pipe.Decl)
		case *tparse.RangeNode:
			fmt.Println(nn.Pipe.Cmds, nn.Pipe.Decl)
		}
	}
}
