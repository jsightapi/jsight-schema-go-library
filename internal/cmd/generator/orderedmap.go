package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// orderedMapGenerator generator will search for `// gen:OrderedMap` comments for
// custom types and generate all necessary code to this type.
//
// Requirements:
//   - Custom type should be a struct with exactly two fields: "data" anf "order";
//   - Field "data" should be a map;
//   - Field "order" should be a slice of map keys;
//   - Field "mx" should a sync.RWMutex.
//
// Known limitations:
//   - Unfortunately "omitempty" tag won't work as expected for ordered maps, so
//     you should add specific code for marshaling.
//
// Added because we should preserve keys order in the map, but Golang don't
// gave to us such ability out of the box. And we want more efficient and clean
// way to manipulate such maps, so empty interface isn't an option.
type orderedMapGenerator struct{}

func (orderedMapGenerator) Name() string { return "OrderedMap" }

func (g orderedMapGenerator) Generate(p string) error {
	pkgName, dd, err := g.parseFile(p)
	if err != nil {
		return fmt.Errorf("faile to parse file: %w", err)
	}

	if err := g.generate(pkgName, dd, filepath.Dir(p)); err != nil {
		return fmt.Errorf("failed to find target types: %w", err)
	}

	return nil
}

func (orderedMapGenerator) parseFile(p string) (pkgName string, dd []ast.Decl, err error) {
	const flags = parser.ParseComments | parser.AllErrors
	f, err := parser.ParseFile(token.NewFileSet(), p, nil, flags)
	if err != nil {
		return "", nil, err
	}

	return f.Name.Name, f.Decls, nil
}

func (g orderedMapGenerator) generate(pkgName string, dd []ast.Decl, dirPath string) error {
	imports := map[string]string{}

	for _, d := range dd {
		decl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		if !g.shouldProcess(decl) {
			continue
		}

		for _, s := range decl.Specs {
			switch spec := s.(type) {
			case *ast.ImportSpec:
				s := strings.Trim(spec.Path.Value, `"`)
				imports[path.Base(s)] = s

			case *ast.TypeSpec:
				om, err := g.collectOrderMap(pkgName, spec, imports)
				if err != nil {
					return fmt.Errorf("failed to process type %q: %w", spec.Name, err)
				}

				if err := g.generateCode(om, dirPath); err != nil {
					return fmt.Errorf("failed to generate code for type %q: %w", om.Name, err)
				}
			}
		}
	}
	return nil
}

type orderedMap struct {
	Name            string
	CapitalizedName string
	PkgName         string
	KeyType         string
	ValueType       string
	UsedImports     map[string]struct{}
}

func (orderedMapGenerator) shouldProcess(d *ast.GenDecl) bool {
	if len(d.Specs) == 0 {
		return false
	}

	if _, ok := d.Specs[0].(*ast.ImportSpec); ok {
		return true
	}

	return strings.Contains(d.Doc.Text(), "gen:OrderedMap")
}

func (g orderedMapGenerator) collectOrderMap(
	pkgName string,
	spec *ast.TypeSpec,
	imports map[string]string,
) (orderedMap, error) {
	strct, ok := spec.Type.(*ast.StructType)
	if !ok {
		return orderedMap{}, nil
	}

	if strct.Fields.NumFields() != 3 {
		return orderedMap{}, errors.New(`OrderedMap should have exactly two fields: "data", "order", and "mutex"`)
	}

	var (
		dataField  *ast.Field
		orderField *ast.Field
		mutexField *ast.Field
	)

	for _, f := range strct.Fields.List {
		switch f.Names[0].Name {
		case "data":
			dataField = f

		case "order":
			orderField = f

		case "mx":
			mutexField = f
		}
	}

	om := orderedMap{
		Name:            spec.Name.Name,
		CapitalizedName: cases.Title(language.English, cases.NoLower).String(spec.Name.Name),
		PkgName:         pkgName,
		UsedImports:     map[string]struct{}{},
	}

	if err := g.checkMutexField(mutexField); err != nil {
		return orderedMap{}, err
	}

	if err := g.collectUsedTypes(dataField, orderField, &om); err != nil {
		return orderedMap{}, err
	}

	if err := g.fillImports(&om, imports); err != nil {
		return orderedMap{}, err
	}

	return om, nil
}

func (orderedMapGenerator) checkMutexField(f *ast.Field) error {
	if f == nil {
		return errors.New(`"mutex" field didn't present'`)
	}

	se, ok := f.Type.(*ast.SelectorExpr)
	if !ok {
		return errors.New(`"mutex" field should be *sync.RWMutex`)
	}

	if x, ok := se.X.(*ast.Ident); !ok || x.Name != "sync" || se.Sel.Name != "RWMutex" {
		return errors.New(`"mutex" field should be *sync.RWMutex`)
	}

	return nil
}

func (orderedMapGenerator) collectUsedTypes(
	data,
	order *ast.Field,
	om *orderedMap,
) error {
	mapType, ok := data.Type.(*ast.MapType)
	if !ok {
		return errors.New(`"data" field should be a map`)
	}

	var err error

	om.KeyType, err = typeToString(mapType.Key)
	if err != nil {
		return fmt.Errorf(`failed to get "data" map key type: %w`, err)
	}

	om.ValueType, err = typeToString(mapType.Value)
	if err != nil {
		return fmt.Errorf(`failed to get "data" map value type: %w`, err)
	}

	slice, ok := order.Type.(*ast.ArrayType)
	if !ok {
		return errors.New(`"order" field should be a slice`)
	}

	sliceType, err := typeToString(slice.Elt)
	if err != nil {
		return fmt.Errorf(`failed to get "order" slice item type: %w`, err)
	}

	if sliceType != om.KeyType {
		return fmt.Errorf(
			`"order" slice item type %q isn't equal to %q`,
			sliceType,
			om.KeyType,
		)
	}
	return nil
}

func (g orderedMapGenerator) fillImports(om *orderedMap, imports map[string]string) error {
	if pkg := g.getTypePackage(om.ValueType); pkg != "" {
		p, ok := imports[pkg]
		if !ok {
			return fmt.Errorf("failed to find import for type %q", om.ValueType)
		}
		om.UsedImports[p] = struct{}{}
	}

	if pkg := g.getTypePackage(om.KeyType); pkg != "" {
		p, ok := imports[pkg]
		if !ok {
			return fmt.Errorf("failed to find import for type %q", om.KeyType)
		}
		om.UsedImports[p] = struct{}{}
	}
	return nil
}

func (orderedMapGenerator) getTypePackage(t string) string {
	parts := strings.SplitN(t, ".", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func (orderedMapGenerator) generateCode(om orderedMap, dirPath string) error {
	t, err := template.New("").Parse(`// Autogenerated code!
// DO NOT EDIT!
//
// Generated by OrderedMap generator from the internal/cmd/generator command.

package {{ .PkgName }}

import (
	"bytes"
	"encoding/json"
{{ range $k, $v := .UsedImports }}
	"{{ $k }}"
{{ end }}
)

// Set sets a value with specified key.
func (m *{{ .Name }}) Set(k {{ .KeyType }}, v {{ .ValueType }}) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.data == nil {
		m.data = map[{{ .KeyType }}]{{ .ValueType }}{}
	}
	if !m.has(k) {
		m.order = append(m.order, k)
	}
	m.data[k] = v
}

// Update updates a value with specified key.
func (m *{{ .Name }}) Update(k {{ .KeyType }}, fn func(v {{ .ValueType }}) {{ .ValueType }}) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if !m.has(k) {
		// Prevent from possible nil pointer dereference if map value type is a
		// pointer.
		return
	}

	m.data[k] = fn(m.data[k])
}

// GetValue gets a value by key.
func (m *{{ .Name }}) GetValue(k {{ .KeyType }}) {{ .ValueType }} {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return m.data[k]
}

// Get gets a value by key.
func (m *{{ .Name }}) Get(k {{ .KeyType }}) ({{ .ValueType }}, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	v, ok := m.data[k]
	return v, ok
}

// Has checks that specified key is set.
func (m *{{ .Name }}) Has(k {{ .KeyType }}) bool {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return m.has(k)
}

func (m *{{ .Name }}) has(k {{ .KeyType }}) bool {
	_, ok := m.data[k]
	return ok
}

// Len returns count of values.
func (m *{{ .Name }}) Len() int {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return len(m.data)
}

func (m *{{ .Name }}) Delete(k {{ .KeyType }}) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.delete(k)
}

func (m *{{ .Name }}) delete(k {{ .KeyType }}) {
var kk {{ .KeyType }}
	i := -1

	for i, kk = range m.order {
		if kk == k {
			break
		}
	}

	delete(m.data, k)
	if i != -1 {
		m.order = append(m.order[:i], m.order[i+1:]...)
	}
}

// Filter iterates and changes values in the map.
func (m *{{ .Name }}) Filter(fn filter{{ .CapitalizedName }}Func) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, k := range m.order {
		if !fn(k, m.data[k]) {
			m.delete(k)
		}
	}
}

type filter{{ .CapitalizedName }}Func = func(k {{ .KeyType }}, v {{ .ValueType }}) bool

// Find finds first matched item from the map.
func (m *{{ .Name }}) Find(fn find{{ .CapitalizedName }}Func) ({{ .Name }}Item, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	for _, k := range m.order {
		if fn(k, m.data[k]) {
			return {{ .Name }}Item{
				Key:   k,
				Value: m.data[k],
			}, true
		}
	}
	return {{ .Name }}Item{}, false
}

type find{{ .CapitalizedName }}Func = func(k {{ .KeyType }}, v {{ .ValueType }}) bool

func (m *{{ .Name }}) Each(fn each{{ .CapitalizedName }}Func) error {
	m.mx.RLock()
	defer m.mx.RUnlock()

	for _, k := range m.order {
		if err := fn(k, m.data[k]); err != nil {
			return err
		}
	}
	return nil
}

type each{{ .CapitalizedName }}Func = func(k {{ .KeyType }}, v {{ .ValueType }}) error

func (m *{{ .Name }}) EachSafe(fn eachSafe{{ .CapitalizedName }}Func) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	for _, k := range m.order {
		fn(k, m.data[k])
	}
}

type eachSafe{{ .CapitalizedName }}Func = func(k {{ .KeyType }}, v {{ .ValueType }})

// Map iterates and changes values in the map.
func (m *{{ .Name }}) Map(fn map{{ .CapitalizedName }}Func) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, k := range m.order {
		v, err := fn(k, m.data[k])
		if err != nil {
			return err
		}
		m.data[k] = v
	}
	return nil
}

type map{{ .CapitalizedName }}Func = func(k {{ .KeyType }}, v {{ .ValueType }}) ({{ .ValueType }}, error)

// {{ .Name }}Item represent single data from the {{ .Name }}.
type {{ .Name }}Item struct {
	Key   {{ .KeyType }}
	Value {{ .ValueType }}
}

var _ json.Marshaler = &{{ .Name }}{}

func (m *{{ .Name }}) MarshalJSON() ([]byte, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	var buf bytes.Buffer
	buf.WriteRune('{')

	for i, k := range m.order {
		if i != 0 {
			buf.WriteRune(',')
		}

		// marshal key
		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteRune(':')

		// marshal value
		val, err := json.Marshal(m.data[k])
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteRune('}')
	return buf.Bytes(), nil
}
`)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	if err = t.Execute(buf, om); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	code, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to gofmt: %w", err)
	}

	p := filepath.Join(dirPath, camelCaseToUnderscore(om.Name)+"_gen.go")
	return os.WriteFile(p, code, 0644) //nolint:gosec // It's okay, we save a code here.
}
