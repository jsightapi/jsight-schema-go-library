package main

import (
	stdBytes "bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// stringerGenerator generator will search for `// gen:Stringer "{some test}"`
// comments for enum and generate implementation of fmt.Stringer.
type stringerGenerator struct{}

func (stringerGenerator) Name() string { return "Stringer" }

func (g stringerGenerator) Generate(path string) error {
	dd, err := g.parseFile(path)
	if err != nil {
		return fmt.Errorf("faile to parse file: %w", err)
	}

	if err := g.generate(dd, filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to find target types: %w", err)
	}

	return nil
}

func (stringerGenerator) parseFile(p string) (dd []ast.Decl, err error) {
	const flags = parser.ParseComments | parser.AllErrors
	f, err := parser.ParseFile(token.NewFileSet(), p, nil, flags)
	if err != nil {
		return nil, err
	}

	return f.Decls, nil
}

func (g stringerGenerator) generate(dd []ast.Decl, dirPath string) error {
	for _, d := range dd {
		decl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		if !g.shouldProcess(decl) {
			continue
		}

		if len(decl.Specs) != 1 {
			continue
		}

		spec, ok := decl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		receiver, comment := g.getReceiverAndComment(decl)

		if err := g.generateCode(spec.Name.Name, receiver, comment, dirPath); err != nil {
			return err
		}
	}
	return nil
}

const stringerMarker = "gen:Stringer"

func (stringerGenerator) shouldProcess(d *ast.GenDecl) bool {
	if len(d.Specs) == 0 {
		return false
	}

	return strings.Contains(d.Doc.Text(), stringerMarker)
}

func (stringerGenerator) getReceiverAndComment(d *ast.GenDecl) (receiver, comment string) {
	if len(d.Specs) == 0 {
		return "", ""
	}

	for _, c := range d.Doc.List {
		idx := strings.Index(c.Text, stringerMarker)
		if idx == -1 {
			continue
		}

		str := strings.TrimSpace(c.Text[idx+len(stringerMarker):])
		parts := strings.SplitN(str, " ", 2)
		if len(parts) != 2 {
			return "", ""
		}
		return parts[0], parts[1]
	}

	return "", ""
}

func (g stringerGenerator) generateCode(typeName, receiver, comment, dirPath string) error {
	outputName := filepath.Join(
		dirPath,
		fmt.Sprintf("%s_string.go", strings.ToLower(buildFileName(typeName))),
	)

	args := []string{
		"-type", typeName,
		"-linecomment",
		"-output", outputName,
	}

	log.Printf(`Run "stringer %s"`, strings.Join(args, " "))
	cmd := exec.Command("stringer", args...)
	cmd.Dir = dirPath

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return g.fixCode(outputName, typeName, receiver, comment)
}

func buildFileName(typeName string) string {
	buf := stdBytes.NewBuffer(make([]byte, 0, len(typeName)))
	for _, r := range typeName {
		if 'A' <= r && r <= 'Z' {
			buf.WriteByte('_')
			r += 'a' - 'A'
		}
		buf.WriteRune(r)
	}
	return strings.Trim(buf.String(), "_")
}

func (stringerGenerator) fixCode(outputName, typeName, receiver, comment string) error {
	content, err := os.ReadFile(outputName)
	if err != nil {
		return fmt.Errorf("read file %q: %w", outputName, err)
	}

	content = bytesReplace(content, `
import "strconv"
`, "")

	content = bytesReplace(
		content,
		fmt.Sprintf(`func (i %[1]s) String() string {
	if i < 0 || i >= %[1]s(len(_%[1]s_index)-1) {
		return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _%[1]s_name[_%[1]s_index[i]:_%[1]s_index[i+1]]
}`, typeName),
		fmt.Sprintf(`func (%[2]s %[1]s) String() string {
	if %[2]s < 0 || %[2]s >= %[1]s(len(_%[1]s_index)-1) {
		panic(%[3]q)
	}
	return _%[1]s_name[_%[1]s_index[%[2]s]:_%[1]s_index[%[2]s+1]]
}`, typeName, receiver, comment),
	)

	content = bytesReplace(
		content,
		fmt.Sprintf(`func (i %[1]s) String() string {
	if i >= %[1]s(len(_%[1]s_index)-1) {
		return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _%[1]s_name[_%[1]s_index[i]:_%[1]s_index[i+1]]
}`, typeName),
		fmt.Sprintf(`func (%[2]s %[1]s) String() string {
	if %[2]s >= %[1]s(len(_%[1]s_index)-1) {
		panic(%[3]q)
	}
	return _%[1]s_name[_%[1]s_index[%[2]s]:_%[1]s_index[%[2]s+1]]
}`, typeName, receiver, comment),
	)

	return os.WriteFile(outputName, content, 0644) //nolint:gosec // It's okay.
}

func bytesReplace(b []byte, olds, news string) []byte {
	return stdBytes.ReplaceAll(b, []byte(olds), []byte(news))
}
