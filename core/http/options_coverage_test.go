package http

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// fiberConfigSkipList records the fiber.Config fields deliberately left without
// a first-class option, and why. They stay reachable through WithFiberConfig.
var fiberConfigSkipList = map[string]string{
	"Services":                        "Fiber's own service lifecycle competes with GOE's fx modules; shipping both would give GOE two lifecycles",
	"ServicesStartupContextProvider":  "part of Fiber's service lifecycle; see Services",
	"ServicesShutdownContextProvider": "part of Fiber's service lifecycle; see Services",
	"RegexHandler":                    "pluggable regex engine; Fiber's own documentation is a better reference than a GOE paraphrase",
}

// listenConfigSkipList records the fiber.ListenConfig fields deliberately left
// without a first-class option, and why. They stay reachable through
// WithListenConfig.
var listenConfigSkipList = map[string]string{
	"GracefulContext": "GOE's shutdown manager owns shutdown sequencing; a second knob would fight goe.Options.ShutdownTimeout",
	"ShutdownTimeout": "superseded by goe.Options.ShutdownTimeout and DrainTimeout",
}

// TestOptionCoverage is the guard that keeps "GOE supports whatever Fiber
// supports" true over time. Every exported field of fiber.Config and
// fiber.ListenConfig must either have a With<Field> constructor or appear on a
// skip list with a stated reason.
//
// When Fiber adds a field, this test fails and names it, so the gap surfaces at
// upgrade time instead of being discovered by a developer who cannot configure
// something a year later.
func TestOptionCoverage(t *testing.T) {
	ctors := optionConstructors(t)
	require.NotEmpty(t, ctors, "no With* option constructors found; the source scan is broken")

	assertFieldsWrapped(t, reflect.TypeFor[fiber.Config](), ctors, fiberConfigSkipList, "fiber.Config")
	assertFieldsWrapped(t, reflect.TypeFor[fiber.ListenConfig](), ctors, listenConfigSkipList, "fiber.ListenConfig")
}

func assertFieldsWrapped(
	t *testing.T,
	typ reflect.Type,
	ctors map[string]bool,
	skip map[string]string,
	label string,
) {
	t.Helper()

	for field := range typ.Fields() {
		if !field.IsExported() {
			continue
		}

		want := "With" + field.Name
		reason, skipped := skip[field.Name]

		switch {
		case skipped && ctors[want]:
			t.Errorf("%s.%s is on the skip list (%s) but %s exists too. "+
				"Remove the skip-list entry.", label, field.Name, reason, want)
		case skipped:
			// Deliberately unwrapped; reachable through the escape hatch.
		case !ctors[want]:
			t.Errorf("%s.%s has no %s option. Add one in options_fiber.go or "+
				"options_listen.go, or add the field to the skip list with a reason.",
				label, field.Name, want)
		}
	}

	// A skip list that names fields Fiber has since removed is misleading, so
	// hold it to the same standard.
	for name, reason := range skip {
		if _, ok := typ.FieldByName(name); !ok {
			t.Errorf("%s skip list names %q (%s), which no longer exists. Remove the entry.",
				label, name, reason)
		}
	}
}

// optionConstructors scans this package's non-test sources for exported
// With* functions. Reading the source rather than maintaining a hand-written
// list is what makes the coverage guarantee self-maintaining.
func optionConstructors(t *testing.T) map[string]bool {
	t.Helper()

	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	fset := token.NewFileSet()
	names := make(map[string]bool)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, name, nil, 0)
		require.NoErrorf(t, err, "parsing %s", name)

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			if strings.HasPrefix(fn.Name.Name, "With") {
				names[fn.Name.Name] = true
			}
		}
	}

	return names
}
