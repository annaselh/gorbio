package core

import "testing"

type stubService struct{ name string }
type otherService struct{}

func TestServiceAsResolvesRegisteredService(t *testing.T) {
	app := NewApp(nil, nil, nil, nil)
	if err := app.Services.Register("stub", &stubService{name: "one"}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	got, err := ServiceAs[*stubService](app, "stub")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if got.name != "one" {
		t.Fatalf("resolved the wrong instance: %q", got.name)
	}
}

func TestServiceAsReportsMissingService(t *testing.T) {
	app := NewApp(nil, nil, nil, nil)

	if _, err := ServiceAs[*stubService](app, "absent"); err == nil {
		t.Fatal("resolving an unregistered name must fail")
	}
}

// Registering under one type and resolving as another is a wiring bug; it must
// surface as an error rather than a nil that panics on first use.
func TestServiceAsRejectsTypeMismatch(t *testing.T) {
	app := NewApp(nil, nil, nil, nil)
	if err := app.Services.Register("stub", &otherService{}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if _, err := ServiceAs[*stubService](app, "stub"); err == nil {
		t.Fatal("a type mismatch must be reported")
	}
}

func TestServiceRegistryRejectsDuplicateName(t *testing.T) {
	registry := NewServiceRegistry()
	if err := registry.Register("stub", &stubService{}); err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}
	if err := registry.Register("stub", &stubService{}); err == nil {
		t.Fatal("registering the same name twice must fail")
	}
}

func TestServiceRegistryRejectsEmptyNameOrNilService(t *testing.T) {
	registry := NewServiceRegistry()
	if err := registry.Register("", &stubService{}); err == nil {
		t.Fatal("a blank service name must be rejected")
	}
	if err := registry.Register("stub", nil); err == nil {
		t.Fatal("a nil service must be rejected")
	}
}
