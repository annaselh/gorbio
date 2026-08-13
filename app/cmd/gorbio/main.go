package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	command := os.Args[1]

	switch command {
	case "module":
		handleModule()

	case "extension":
		handleExtension()

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		usage()
	}
}

func usage() {
	fmt.Print(`
Gorbio CLI

Usage:

  gorbio module create <name>
  gorbio extension create <name>

Examples:

  gorbio module create inventory
  gorbio extension create sales-discount
`)
}

func handleModule() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gorbio module create <name>")
		return
	}

	action := os.Args[2]
	name := normalizeName(os.Args[3])

	switch action {
	case "create":
		if err := createModule(name); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown module action:", action)
	}
}

func handleExtension() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gorbio extension create <name>")
		return
	}

	action := os.Args[2]
	name := normalizeName(os.Args[3])

	switch action {
	case "create":
		if err := createExtension(name); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown extension action:", action)
	}
}

func createModule(name string) error {
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	moduleDir := filepath.Join("modules", name)

	if _, err := os.Stat(moduleDir); err == nil {
		return fmt.Errorf("module %q already exists", name)
	}

	dirs := []string{
		moduleDir,
		filepath.Join(moduleDir, "migrations"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"module.go": moduleTemplate(name),

		"models.go": fmt.Sprintf(
			"package %s\n",
			name,
		),

		"service.go": fmt.Sprintf(
			"package %s\n",
			name,
		),

		"handlers.go": fmt.Sprintf(
			"package %s\n",
			name,
		),

		"routes.go": fmt.Sprintf(
			"package %s\n",
			name,
		),

		"permissions.go": fmt.Sprintf(
			"package %s\n",
			name,
		),

		"migrations/001_init.go": `package migrations
`,
	}

	for file, content := range files {
		path := filepath.Join(moduleDir, file)

		if err := os.WriteFile(
			path,
			[]byte(content),
			0644,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Module %q created successfully\n", name)
	fmt.Println()
	fmt.Printf("  Location: %s\n", moduleDir)
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  go generate ./modules")

	return nil
}

func createExtension(name string) error {
	if name == "" {
		return fmt.Errorf("extension name cannot be empty")
	}

	extensionDir := filepath.Join("extensions", name)

	if _, err := os.Stat(extensionDir); err == nil {
		return fmt.Errorf("extension %q already exists", name)
	}

	pkg := packageName(name)

	dirs := []string{
		extensionDir,
		filepath.Join(extensionDir, "migrations"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"extension.go": extensionTemplate(name, pkg),

		"models.go": fmt.Sprintf(
			"package %s\n",
			pkg,
		),

		"service.go": fmt.Sprintf(
			"package %s\n",
			pkg,
		),

		"handlers.go": fmt.Sprintf(
			"package %s\n",
			pkg,
		),

		"migrations/001_init.go": `package migrations
`,
	}

	for file, content := range files {
		path := filepath.Join(extensionDir, file)

		if err := os.WriteFile(
			path,
			[]byte(content),
			0644,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Extension %q created successfully\n", name)
	fmt.Println()
	fmt.Printf("  Location: %s\n", extensionDir)
	fmt.Printf("  Package:  %s\n", pkg)
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  go generate ./extensions")

	return nil
}

func moduleTemplate(name string) string {
	return fmt.Sprintf(`package %s

import "github.com/annaselh/gorbio/core"

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "%s"
}

func (m *Module) Depends() []string {
	return []string{
		"base",
	}
}

func (m *Module) Register(app *core.App) error {
	return nil
}

func (m *Module) Boot(app *core.App) error {
	return nil
}
`, name, name)
}

func extensionTemplate(name, pkg string) string {
	return fmt.Sprintf(`package %s

import "github.com/annaselh/gorbio/core"

type Extension struct{}

func New() *Extension {
	return &Extension{}
}

func (e *Extension) Name() string {
	return "%s"
}

func (e *Extension) Module() string {
	return ""
}

func (e *Extension) Depends() []string {
	return []string{}
}

func (e *Extension) Register(app *core.App) error {
	return nil
}

func (e *Extension) Boot(app *core.App) error {
	return nil
}
`, pkg, name)
}

func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	var result strings.Builder

	for _, r := range name {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			r == '-' ||
			r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func packageName(name string) string {
	name = strings.ToLower(name)

	var result strings.Builder

	for _, r := range name {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}
