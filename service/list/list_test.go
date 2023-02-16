package list

import (
	"path"
	"testing"
)

const (
	Root = "/var/log"
)

func buildProperties(name string) *properties {
	props := new(properties)
	props.name = name
	props.rootedPath = path.Join(Root, name)
	return props
}

func TestExtractParams(t *testing.T) {
	// TODO:
	// Need to construct/mock the HTTP request.
}

func TestListDir_nilFilter(t *testing.T) {
	// TODO:
	// This test should exercise listDir().
	// Need to mock file system.
}

func TestListDir_negFilter(t *testing.T) {
	// TODO:
	// This test should exercise listDir().
	// Need to mock file system.
}

func TestListDir_posFilter(t *testing.T) {
	// TODO:
	// This test should exercise listDir().
	// Need to mock file system.
}

func TestListFile_nilFilter(t *testing.T) {
	props := buildProperties("name")
	data, err := props.listFile()
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 1 {
		t.Errorf("expected data len 1, got %d\n", len(data))
	}
	if data[0].Name != "/var/log/name" {
		t.Errorf("Expected data.Name '/var/log/name', got %v", data[0].Name)
	}
}

func TestListFile_negFilter(t *testing.T) {
	props := buildProperties("name")
	props.filterText = "n"
	props.filterOmit = true
	data, err := props.listFile()
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 0 {
		t.Errorf("expected data len 0, got %d\n", len(data))
	}

	props.filterText = "z"
	data, err = props.listFile()
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 1 {
		t.Errorf("expected data len 1, got %d", len(data))
	}
	if data[0].Name != "/var/log/name" {
		t.Errorf("Expected data.Name '/var/log/name', got %v", data[0].Name)
	}
}

func TestListFile_posFilter(t *testing.T) {
	props := buildProperties("name")
	props.filterText = "z"
	props.filterOmit = false
	data, err := props.listFile()
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 0 {
		t.Errorf("expected data len 0, got %d\n", len(data))
	}

	props.filterText = "a"
	data, err = props.listFile()
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 1 {
		t.Errorf("expected data len 1, got %d", len(data))
	}
	if data[0].Name != "/var/log/name" {
		t.Errorf("Expected data.Name '/var/log/name', got %v", data[0].Name)
	}
}

func TestStripRootPrefix(t *testing.T) {
	var m metadata
	m = metadata{Name: "/var/log/abc"}
	m.stripRootPrefix("/var/log/")
	if m.Name != "abc" {
		t.Errorf("Expected 'abc', got %q\n", m.Name)
	}
	m = metadata{Name: "/other/root/a/b/c"}
	m.stripRootPrefix("/other/root/")
	if m.Name != "a/b/c" {
		t.Errorf("Expected 'a/b/c', got %q\n", m.Name)
	}
}

func TestValidateParams(t *testing.T) {
	props := buildProperties("a/b/c")
	err := props.validateParams()
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	expected := Root + "/" + "a/b/c"
	if props.rootedPath != expected {
		t.Errorf("Expected path %q, got %q", expected, props.rootedPath)
	}

	props.name = "//./x/./y"
	err = props.validateParams()
	if err != nil {
		t.Errorf("Expected error but got %v", err)
	}
	expected = Root + "/" + "x/y"
	if props.rootedPath != expected {
		t.Errorf("Expected path %q, got %q", expected, props.rootedPath)
	}

	props.name = "../../x/y"
	err = props.validateParams()
	if err == nil {
		t.Errorf("Expected error but got nil, path %q", props.rootedPath)
	}
}
