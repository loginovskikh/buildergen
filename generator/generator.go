package generator

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Generator struct {
	SourceFile     string
	ParsedImports  map[string]string
	NeededImports  []string
	Fields         map[string]string
	TypeName       string
	OutputFileName string
	PackageName    string
	Ast            *ast.File
}

func NewGenerator(sourceFilePath, typeName, outputFileName string) Generator {
	return Generator{
		ParsedImports:  make(map[string]string, 0),
		NeededImports:  make([]string, 0, 1),
		Fields:         make(map[string]string, 0),
		TypeName:       typeName,
		OutputFileName: outputFileName,
		SourceFile:     sourceFilePath,
	}
}

func (g *Generator) Generate() error {
	err := g.getAst()
	if err != nil {
		return err
	}
	err = g.parseAst()
	if err != nil {
		return err
	}

	err = g.generateOutput()

	return err
}

func (g *Generator) generateOutput() error {
	builderName := g.TypeName + "Builder"

	output := fmt.Sprintf("package %s\n\n", g.PackageName)

	for _, fieldType := range g.NeededImports {
		if importString, ok := g.ParsedImports[fieldType]; ok {
			output += fmt.Sprintf(`import %s
`, importString)
		}
	}

	output += fmt.Sprintf(`
type %s struct {
	instance *%s
}
		
func %sBuild() *%s {
	return &%s {
		instance: &%s{},
	}
}`, builderName, g.TypeName, g.TypeName, builderName, builderName, g.TypeName)

	caser := cases.Title(language.English)

	for fieldName, fieldType := range g.Fields {
		output += fmt.Sprintf(`
func (b *%s) %s (%s %s) *%s {
	b.instance.%s = %s
	return b
}
`, builderName, caser.String(fieldName), fieldName, fieldType, builderName, fieldName, fieldName)
	}

	output += fmt.Sprintf(`
func (b *%s) P() *%s {
	return b.instance
}
	`, builderName, g.TypeName)

	output += fmt.Sprintf(`
func (b *%s) V() %s {
	return *b.instance
}
`, builderName, g.TypeName)

	err := g.save(output)

	return err
}

func (g *Generator) save(output string) error {
	packagePath := filepath.Dir(g.SourceFile)
	err := os.MkdirAll(packagePath, os.ModePerm)

	outputFilePath := filepath.Join(packagePath, g.OutputFileName)
	outputFile, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		if err != nil {
			return err
		}
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(output)

	return err
}

func (g *Generator) getAst() error {
	fset := token.NewFileSet()
	astf, err := parser.ParseFile(fset, g.SourceFile, nil, 0)
	if err != nil {
		return err
	}

	if astf.Name.Name == "" {
		return errors.New("invalid input file: name of package is not specified")
	}

	g.Ast = astf

	return nil
}

func (g *Generator) parseAst() error {

	g.PackageName = g.Ast.Name.Name
	defFound := false
	for _, decl := range g.Ast.Decls {
		declObj, ok := decl.(*ast.GenDecl)
		if ok {
			if declObj.Tok.String() == "import" {
				for _, spec := range declObj.Specs {
					importSpec, ok := spec.(*ast.ImportSpec)
					if ok {
						importName, importPath := g.getImport(importSpec)
						if importName == "" {
							importName = importPath
						} else {
							importPath = importName + " " + importPath
						}
						importName = strings.Trim(importName, `"`)
						g.ParsedImports[importName] = importPath
					}
				}
			}
			if declObj.Tok.String() == "type" {
				for _, spec := range declObj.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if ok {
						if typeSpec.Name.Name == g.TypeName {
							defFound = true
							structType, ok := typeSpec.Type.(*ast.StructType)
							if !ok {
								return errors.New(g.TypeName + "is not a struct")
							}
							for _, field := range structType.Fields.List {
								g.parseFieldData(*field)
							}
						}
					}
				}
			}
		}
		if defFound {
			break
		}
	}

	if !defFound {
		return errors.New("type definition not found")
	}
	if len(g.Fields) == 0 {
		return errors.New("empty structure type")
	}

	return nil
}

func (g *Generator) parseFieldData(field ast.Field) {
	fieldType := g.getFieldTypeString(field.Type)
	fieldNames := g.getFieldNames(field)

	for _, name := range fieldNames {
		g.Fields[name] = fieldType
	}
}

func (g *Generator) getFieldTypeString(typeDef ast.Expr) string {
	typeString := ""
	switch typeDef.(type) {
	case *ast.StarExpr:
		t := typeDef.(*ast.StarExpr)
		typeString += "*" + g.getFieldTypeString(t.X)

	case *ast.SelectorExpr:
		t := typeDef.(*ast.SelectorExpr)
		packageName := g.getFieldTypeString(t.X)
		identName := g.getFieldTypeString(t.Sel)
		g.NeededImports = append(g.NeededImports, packageName)
		typeString += packageName + "." + identName

	case *ast.FuncType:
		typeString = "func("
		t := typeDef.(*ast.FuncType)
		typeString = g.formatFuncTypeString(t)

	case *ast.Ident:
		t := typeDef.(*ast.Ident)
		typeString = t.Name
	}

	return typeString
}

func (g Generator) formatFuncTypeString(t *ast.FuncType) string {
	typeItems := make([]string, 0, 5)

	paramsItems := make([]string, 0, len(t.Params.List))

	for _, param := range t.Params.List {
		params := g.getFieldNames(*param)
		paramsType := g.getFieldTypeString(param.Type)
		paramsItems = append(paramsItems, strings.Join(params, ",")+" "+paramsType)
	}
	paramsItemsString := strings.Join(paramsItems, ", ")

	returnItems := make([]string, 0, len(t.Params.List))

	for _, result := range t.Results.List {
		resultNames := g.getFieldNames(*result)
		resultTypes := g.getFieldTypeString(result.Type)
		returnItems = append(returnItems, strings.Join(resultNames, ",")+" "+resultTypes)
	}
	returnItemsString := strings.Join(returnItems, ", ")

	if len(returnItems) > 1 || strings.Contains(returnItemsString, " ") {
		returnItemsString = "(" + returnItemsString + ")"
	}

	typeItems = append(typeItems, "func(")
	typeItems = append(typeItems, paramsItemsString)
	typeItems = append(typeItems, ")")
	typeItems = append(typeItems, returnItemsString)

	return strings.Join(typeItems, " ")
}

func (g Generator) getFieldNames(field ast.Field) []string {
	names := make([]string, 0, len(field.Names))
	for _, name := range field.Names {
		names = append(names, name.Name)
	}

	return names
}

func (g Generator) getImport(importSpec *ast.ImportSpec) (string, string) {
	name := ""
	importString := ""

	if importSpec.Name != nil {
		name = importSpec.Name.Name
	}

	if importSpec.Path != nil {
		importString = importSpec.Path.Value
	}
	if name == "_" {
		name = ""
	}

	return name, importString
}
