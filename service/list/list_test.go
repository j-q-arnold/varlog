package list


import (
	"path"
	"testing"
)

const (
	Root = "/var/log"
)

func buildParams(name string) *properties {
	params := new(properties)
	params.name = name
	params.rootedPath = path.Join(Root, name)
	return params
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
	params := buildParams("name")
	data, err := listFile(params)
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 1 {
		t.Errorf("expected data len 1, got %d\n", len(data));
	}
	if data[0].Name != "/var/log/name" {
		t.Errorf("Expected data.Name '/var/log/name', got %v", data[0].Name)
	}
}

func TestListFile_negFilter(t *testing.T) {
	params := buildParams("name")
	params.filterText = "n"
	params.filterOmit = true
	data, err := listFile(params)
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 0 {
		t.Errorf("expected data len 0, got %d\n", len(data))
	}

	params.filterText = "z"
	data, err = listFile(params)
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
	params := buildParams("name")
	params.filterText = "z"
	params.filterOmit = false
	data, err := listFile(params)
	if err != nil {
		t.Errorf("expected nil error, got %v\n", err)
	}
	if len(data) != 0 {
		t.Errorf("expected data len 0, got %d\n", len(data))
	}

	params.filterText = "a"
	data, err = listFile(params)
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
	if (m.Name != "abc") {
		t.Errorf("Expected 'abc', got %q\n", m.Name)
	}
	m = metadata{Name: "/other/root/a/b/c"}
	m.stripRootPrefix("/other/root/")
	if (m.Name != "a/b/c") {
		t.Errorf("Expected 'a/b/c', got %q\n", m.Name);
	}
}

func TestValidateParams(t *testing.T) {
	params := buildParams("a/b/c")
	err := validateParams(params)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	expected := Root + "/" + "a/b/c"
	if params.rootedPath != expected {
		t.Errorf("Expected path %q, got %q", expected, params.rootedPath)
	}

	params.name = "//./x/./y"
	err = validateParams(params)
	if err != nil {
		t.Errorf("Expected error but got %v", err)
	}
	expected = Root + "/" + "x/y"
	if params.rootedPath != expected {
		t.Errorf("Expected path %q, got %q", expected, params.rootedPath)
	}

	params.name = "../../x/y"
	err = validateParams(params)
	if err == nil {
		t.Errorf("Expected error but got nil, path %q", params.rootedPath)
	}
}