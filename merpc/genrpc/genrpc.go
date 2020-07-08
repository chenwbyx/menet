//go:generate go install menet/merpc/genrpc
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

func importName(path string) string {
	paths := strings.Split(path[1:len(path)-1], "/")
	return paths[len(paths)-1]
}

type RpcVisitor struct {
	rpcNames  []string
	services  map[string][]*ast.FuncDecl
	importers map[string]string
}

func newXRpcVisitor() *RpcVisitor {
	x := &RpcVisitor{}
	x.services = make(map[string][]*ast.FuncDecl)
	x.importers = make(map[string]string)
	return x
}

type rpcService struct {
	name string
}

func (v *RpcVisitor) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.FuncDecl:
		Func := node
		if Func.Recv != nil {
			if len(Func.Type.Params.List) == 2 && Func.Type.Results != nil && len(Func.Type.Results.List) == 1 {

				reqType, ok1 := Func.Type.Params.List[0].Type.(*ast.StarExpr)
				respType, ok2 := Func.Type.Params.List[1].Type.(*ast.StarExpr)
				errType, ok := Func.Type.Results.List[0].Type.(*ast.Ident)
				if ok1 && ok2 && ok && errType.Name == "error" {
					fmt.Println("Type", Func.Recv.List[0].Type, Func.Name.Name)
					//req := reqType.X.(*ast.Ident).Name
					//resp := respType.X.(*ast.Ident).Name
					_, _ = reqType, respType
					switch t := Func.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						ident, ok := t.X.(*ast.Ident)
						if ok {
							v.services[ident.Name] = append(v.services[ident.Name], Func)
						}
					case *ast.Ident:
						v.services[t.Name] = append(v.services[t.Name], Func)
					}

				}
			}
		}
	case *ast.Comment:
		comment := strings.TrimSpace(strings.TrimPrefix(node.Text, "//"))

		if strings.HasPrefix(comment, "rpc:") {
			for i, name := range strings.Split(comment[4:], " ") {
				if name != "" {
					fmt.Println(i, name)
					v.rpcNames = append(v.rpcNames, strings.TrimSpace(name))
				}
			}
		}
	case *ast.TypeSpec:
	case *ast.ImportSpec:
		Import := node
		if Import.Name != nil {
			v.importers[Import.Name.Name] = Import.Path.Value
		} else {
			path := Import.Path.Value[1 : len(Import.Path.Value)-1]
			pkgName := getPkgName(path)
			if pkgName != "" {
				v.importers[pkgName] = Import.Path.Value
			} else {
				log.Println("no pkg name", Import.Path.Value)
			}
		}
	}
	return v
}

func (v *RpcVisitor) Gen() {
	goPath := os.Getenv("GOPATH")
	cwd, _ := os.Getwd()
	for _, pkgName := range v.rpcNames {
		rpcMethods, has := v.services[pkgName]
		if !has {
			continue
		}
		var genPath string
		if cwd != "" {
			genPath = filepath.Join(cwd, "rpc_client", pkgName)
		} else {
			genPath = filepath.Join(goPath, "src/rpc_client", pkgName)
		}
		err := os.MkdirAll(genPath, 0770)
		if err != nil {
			fmt.Println("dir", genPath, err)
			continue
		}
		f, err := os.OpenFile(filepath.Join(genPath, pkgName+".go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
		if err != nil {
			fmt.Println("file", filepath.Join(genPath, pkgName+".go"), err)
			continue
		}

		f.Write([]byte("package " + pkgName + "\n\n"))
		var imps = make(map[string]Import)
		for _, meth := range rpcMethods {
			for i := range []int32{0, 1} {
				m, _ := getSelectorExpr(meth.Type.Params.List[i].Type.(*ast.StarExpr))
				if m != "" {
					imps[v.importers[m]] = Import{Alias: m, Path: v.importers[m]}
				}
			}
		}
		{
			importT := template.Must(template.New("imports").Parse(importTmpl))
			importT.Execute(f, imps)
			initT := template.Must(template.New("init").Parse(initTmpl))
			initT.Execute(f, pkgName)
		}
		funcMaps := template.FuncMap{"lower": strings.ToLower}
		t := template.Must(template.New("method").Funcs(funcMaps).Parse(methodTmpl))
		for _, meth := range rpcMethods {
			t.Execute(f, &Meth{
				PkgName:  pkgName,
				FuncName: meth.Name.Name,
				Req:      extraName(meth.Type.Params.List[0].Type),
				Resp:     extraName(meth.Type.Params.List[1].Type),
			})
		}

		f.Close()
	}
}

func getPkgName(path string) string {
	goPath := os.Getenv("GOPATH")
	goRoot := os.Getenv("GOROOT")
	pkgPath := filepath.Join(goPath, "src", path)
	dPkgPath := filepath.Join(goRoot, "src", path)
	if _, e := os.Stat(pkgPath); !os.IsNotExist(e) {
		fs := token.NewFileSet()
		pkgs, err := parser.ParseDir(fs, pkgPath, nil, parser.PackageClauseOnly)
		if err != nil {
			log.Println(err.Error())
			return ""
		}
		for name := range pkgs {
			if strings.HasSuffix(name, "_test") == false {
				return name
			}
		}
	}
	if _, e := os.Stat(dPkgPath); !os.IsNotExist(e) {
		fs := token.NewFileSet()
		pkgs, err := parser.ParseDir(fs, dPkgPath, nil, parser.PackageClauseOnly)
		if err != nil {
			log.Println(err.Error())
			return ""
		}
		for name := range pkgs {
			if strings.HasSuffix(name, "_test") == false {
				return name
			}
		}
	}

	return ""
}

const importTmpl = `
import (
	"menet/merpc"
	"sync"
	{{range .}}{{if (eq "" .Alias)}}{{.Path}}{{else}}{{.Alias}} {{.Path}}{{end}}{{print "\n"}}{{end}})
`

const initTmpl = `
type {{.}}Client struct {
	c *merpc.RpcClient
}

func New(addr string) *{{.}}Client {
	c := &{{.}}Client{}
	c.c = merpc.NewClient(addr)
	return c
}

func (c *{{.}}Client)Close() {
	c.c.Close()
}

var g{{.}}Client *{{.}}Client
var once{{.}} sync.Once
func Init(addr string) {
	once{{.}}.Do(func() {
		g{{.}}Client = New(addr)
	})
}

func GetClient() *{{.}}Client {
	return g{{.}}Client
}
`
const methodTmpl = `
type I{{.FuncName}} interface {
	{{.FuncName}}(*{{.Req}}, *{{.Resp}}) error
}
func (c *{{.PkgName}}Client){{.FuncName}}(req *{{.Req}}) (*{{.Resp}}, error) {
	var resp {{.Resp}}
	err := c.c.Call("{{.PkgName}}.{{.FuncName}}", &req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
func (c *{{.PkgName}}Client){{.FuncName}}Ex(req *{{.Req}}, resp *{{.Resp}}) error {
	err := c.c.Call("{{.PkgName}}.{{.FuncName}}", req, resp)
	if err != nil {
		return err
	}
	return nil
}

func {{.FuncName}}(req *{{.Req}}) (*{{.Resp}}, error) {
	return g{{.PkgName}}Client.{{.FuncName}}(req)
}
func {{.FuncName}}Ex(req *{{.Req}}, resp *{{.Resp}}) error {
	return g{{.PkgName}}Client.{{.FuncName}}Ex(req, resp)
}
`

type Meth struct {
	PkgName  string
	FuncName string
	Req      string
	Resp     string
}

type Import struct {
	Alias   string
	PkgName string
	Path    string
}

func getSelectorExpr(expr *ast.StarExpr) (m, sel string) {
	switch v := expr.X.(type) {
	case *ast.Ident:
		return "", v.Name
	case *ast.SelectorExpr:
		return v.X.(*ast.Ident).Name, v.Sel.Name
	default:
		return "", ""
	}
}

var nameCount int

func extraName(expr ast.Expr) string {
	switch t := expr.(*ast.StarExpr).X.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", t.X.(*ast.Ident).Name, t.Sel.Name)
	default:
		nameCount += 1
		return "Unknown" + strconv.Itoa(nameCount)
	}
}

func main() {
	goFile := os.Getenv("GOFILE")
	goPackage := os.Getenv("GOPACKAGE")

	goPath := os.Getenv("GOPATH")

	cwd, _ := os.Getwd()

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, cwd, nil, parser.AllErrors|parser.ParseComments)
	//f, err := parser.ParseFile(fs, goFile, nil, parser.AllErrors)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var pkg *ast.Package
	for _, p := range pkgs {
		if strings.HasSuffix(p.Name, "_test") == false {
			pkg = p
			break
		}
	}
	if pkg == nil {
		return
	}

	fmt.Println(goFile, goPackage, goPath)
	fmt.Println(cwd)
	//ast.Print(fs, pkgs)

	rpcGen := newXRpcVisitor()
	rpcGen.services = make(map[string][]*ast.FuncDecl)
	ast.Walk(rpcGen, pkg)
	rpcGen.Gen()
}
