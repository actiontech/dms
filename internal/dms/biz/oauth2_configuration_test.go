package biz

import "testing"

func TestParseJsonByPath(t *testing.T) {

	// Sample JSON
	json := `{"name":"John", "age":30, "Hobbies":["football","tennis"], "children":[{"Name":"Emma","age":5},{"Name":"Michael","age":2}]}`

	// Test parsing simple field
	value, err := ParseJsonByPath([]byte(json), "name")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value.ToString() != "John" {
		t.Errorf("Expected name to be John, got %s", value.ToString())
	}

	// Test parsing array element
	value, err = ParseJsonByPath([]byte(json), "Hobbies.0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value.ToString() != "football" {
		t.Errorf("Expected first hobby to be football, got %s", value.ToString())
	}

	// Test parsing nested object
	value, err = ParseJsonByPath([]byte(json), "children.0.Name")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value.ToString() != "Emma" {
		t.Errorf("Expected first child name to be Emma, got %s", value.ToString())
	}

	// Test invalid path
	_, err = ParseJsonByPath([]byte(json), "invalidField")
	if err == nil {
		t.Error("Expected error for invalid path")
	}

	// Test empty path
	_, err = ParseJsonByPath([]byte(json), "")
	if err == nil {
		t.Error("Expected error for empty path")
	}

}
