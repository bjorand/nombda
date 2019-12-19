package engine

import (
	"testing"
)

func setup() {
	runs = make(map[string]*Run, 1024)
}

func TestHookOnlyIf(t *testing.T) {
	setup()

	h, err := ReadHook("tests/hooks", "tests", "test_only_if")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.asyncRun(r)
	output := r.Registers["foo"]
	expected := ""
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["bar"]
	expected = "bar"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookRegister(t *testing.T) {
	setup()

	h, err := ReadHook("tests/hooks", "tests", "test_register")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.asyncRun(r)
	output := r.Registers["foo"]
	expected := "bar"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["bar"]
	expected = "foo"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookHandler(t *testing.T) {
	setup()

	h, err := ReadHook("tests/hooks", "tests", "test_handler")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.asyncRun(r)
	output := r.Registers["foo"]
	expected := "bar"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["bar"]
	expected = "foo"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookCommandFailure(t *testing.T) {
	setup()

	h, err := ReadHook("tests/hooks", "tests", "test_command_failure")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.asyncRun(r)
	outputInt := r.ExitCode
	expectedInt := 2
	if outputInt != expectedInt {
		t.Fatalf("want %+v, got %+v", expectedInt, outputInt)
	}
	output := r.Registers["bar"]
	expected := "foo"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["foo"]
	expected = ""
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookHandlerVars(t *testing.T) {
	setup()

	h, err := ReadHook("tests/hooks", "tests", "test_handler_vars")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.asyncRun(r)
	outputInt := r.ExitCode
	expectedInt := 0
	if outputInt != expectedInt {
		t.Fatalf("want %+v, got %+v", expectedInt, outputInt)
	}
	output := r.Registers["arg_a"]
	expected := "1"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["arg_b"]
	expected = "2"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["arg_c"]
	expected = "3"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}
