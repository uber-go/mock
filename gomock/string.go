package gomock

import "fmt"

type mockInstance interface {
	ISGOMOCK() struct{}
}
type mockedStringer interface {
	fmt.Stringer
	mockInstance
}

// getString is a safe way to convert a value to a string for printing results
// If the value is a a mock, getString avoids calling the mocked String() method,
// which avoids potential deadlocks
func getString(x any) string {
	switch v := x.(type) {
	case mockedStringer:
		return fmt.Sprintf("%T", v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
