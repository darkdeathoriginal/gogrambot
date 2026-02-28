package helpers

import (
	"testing"
)

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := TestStruct{Name: "John", Age: 30}
	result := StructToMap(data)

	if result == nil {
		t.Fatalf("Expected map, got nil")
	}

	if result["name"] != "John" {
		t.Errorf("Expected name to be 'John', got '%v'", result["name"])
	}

	if age, ok := result["age"].(float64); !ok || age != 30 {
		t.Errorf("Expected age to be 30, got '%v'", result["age"])
	}

	invalidData := make(chan int)
	resultInvalid := StructToMap(invalidData)
	if resultInvalid != nil {
		t.Errorf("Expected nil for invalid data, got '%v'", resultInvalid)
	}
}
