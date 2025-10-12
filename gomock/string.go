package gomock

import (
	"fmt"
	"reflect"
)

// getString is a safe way to convert a value to a string for printing results
// If the value is a a mock, getString avoids calling the mocked String() method,
// which avoids potential deadlocks
func getString(x any) string {
	if isGeneratedMock(x) {
		return fmt.Sprintf("%T", x)
	}
	typ := reflect.ValueOf(x)
	if typ.Kind() == reflect.Ptr && typ.IsNil() {
		return "nil"
	}
	if s, ok := x.(fmt.Stringer); ok {
		// Use defer/recover to handle panics from calling String() on nil receivers
		// This matches the behavior of fmt.Sprintf("%v", x) which handles nil Stringers safely
		return safeString(s)
	}
	return fmt.Sprintf("%v", x)
}

// safeString calls String() with panic recovery to handle nil receivers
func safeString(s fmt.Stringer) (result string) {
	defer func() {
		if r := recover(); r != nil {
			// If String() panicked (e.g., nil receiver), use fmt.Sprintf instead
			result = fmt.Sprintf("%v", s)
		}
	}()
	return s.String()
}

// isGeneratedMock checks if the given type has a "isgomock" field,
// indicating it is a generated mock.
func isGeneratedMock(x any) bool {
	typ := reflect.TypeOf(x)
	if typ == nil {
		return false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	_, isgomock := typ.FieldByName("isgomock")
	return isgomock
}
