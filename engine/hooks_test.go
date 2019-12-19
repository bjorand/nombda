package engine

import (
	"strings"
	"testing"
)

func TestHookOnlyIf(t *testing.T) {
	h, err := ReadHook("tests/hooks", "tests", "test_only_if")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
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
	h, err := ReadHook("tests/hooks", "tests", "test_register")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
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
	h, err := ReadHook("tests/hooks", "tests", "test_handler")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
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
	h, err := ReadHook("tests/hooks", "tests", "test_command_continue_after_failure")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
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
	h, err := ReadHook("tests/hooks", "tests", "test_handler_vars")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
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

func TestHookHandlerOnFailure(t *testing.T) {
	h, err := ReadHook("tests/hooks", "tests", "test_handler_on_failure")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
	outputInt := r.ExitCode
	expectedInt := 0
	if outputInt != expectedInt {
		t.Fatalf("want %+v, got %+v", expectedInt, outputInt)
	}
	output := r.Registers["bar"]
	expected := "foo"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookHandlerOnFailureFails(t *testing.T) {
	h, err := ReadHook("tests/hooks", "tests", "test_handler_on_failure_fails")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
	outputInt := r.ExitCode
	expectedInt := 1
	if outputInt != expectedInt {
		t.Fatalf("want %+v, got %+v", expectedInt, outputInt)
	}
	output := r.Registers["bar"]
	expected := ""
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookHandlerOnFailureContinueAfterFailure(t *testing.T) {
	h, err := ReadHook("tests/hooks", "tests", "test_handler_on_failure_continue_after_failure")
	if err != nil {
		t.Fatal(err)
	}
	outputB := h.Tasks[0].ContinueAfterFailure
	expectedB := true
	if outputB != expectedB {
		t.Fatalf("want %+v, got %+v", expectedB, outputB)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
	outputInt := r.ExitCode
	expectedInt := 0
	if outputInt != expectedInt {
		t.Fatalf("want %+v, got %+v", expectedInt, outputInt)
	}
	output := r.Registers["a"]
	expected := "1"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = r.Registers["bar"]
	expected = "foo"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}

func TestHookCommandSecret(t *testing.T) {
	h, err := ReadHook("tests/hooks", "tests", "test_command_secret")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}
	r.Secrets["foo"] = "123"
	h.AsyncRun(r)
	output := r.Registers["foo"]
	expected := "secret:123"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	for _, line := range strings.Split(r.Log(), "\n") {
		if strings.HasPrefix(line, "secret:") {
			output := line
			expected := "secret:***"
			if output != expected {
				t.Fatalf("want %+v, got %+v", expected, output)
			}
		}
	}
}

func TestHookInjectSecret(t *testing.T) {
	h, err := ReadHook("tests/hooks", "tests", "test_command_secret")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(h)
	if err != nil {
		t.Fatal(err)
	}

	r.Secrets, err = ReadSecretFile("tests/secrets/secrets.yml")
	if err != nil {
		t.Fatal(err)
	}
	h.AsyncRun(r)
	output := r.Registers["foo"]
	expected := "secret:bar"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	for _, line := range strings.Split(r.Log(), "\n") {
		if strings.HasPrefix(line, "secret:") {
			output := line
			expected := "secret:***"
			if output != expected {
				t.Fatalf("want %+v, got %+v", expected, output)
			}
		}
	}
}
