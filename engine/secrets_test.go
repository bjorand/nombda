package engine

import "testing"

func TestReadSecretFile(t *testing.T) {
	secrets, err := ReadSecretFile("tests/secrets/secrets.yml")
	if err != nil {
		t.Fatal(err)
	}
	output := secrets["foo"]
	expected := "bar"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
	output = secrets["a"]
	expected = "1"
	if output != expected {
		t.Fatalf("want %+v, got %+v", expected, output)
	}
}
