package resolver

import (
	"fmt"
	"reflect"
)

func ValidateStructFields(expectedFields map[string]reflect.Type, structType reflect.Type) error {
	fieldCount := structType.NumField()

	// Check if the number of fields matches the expected count
	if fieldCount != len(expectedFields) {
		return fmt.Errorf("Expected %d fields, but got %d fields", len(expectedFields), fieldCount)
	}

	// Check if the field types match the expected types
	for i := 0; i < fieldCount; i++ {
		field := structType.Field(i)
		expectedType, ok := expectedFields[field.Name]
		if !ok {
			return fmt.Errorf("Unexpected field: %s", field.Name)
		}
		if field.Type != expectedType {
			return fmt.Errorf("Field %s: expected type %v, but got %v", field.Name, expectedType, field.Type)
		}
	}
	return nil
}
