package utils

import (
	"reflect"
)

// CheckIfPointer checks if the given value / interface is/has a pointer.
func CheckIfPointer(val any) bool {
	// Get the reflection value of the interface
	reflectVal := reflect.ValueOf(val)

	// Check if the value itself is nil
	if !reflectVal.IsValid() {
		//return errors.New("value is nil")
		return false
	}

	// Get the type of the reflected value
	reflectType := reflectVal.Type()

	// Check if the type is a pointer
	if reflectType.Kind() != reflect.Ptr {
		//return errors.New("value must be a pointer")
		return false
	}

	return true
}
