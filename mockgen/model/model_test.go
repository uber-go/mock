package model

import (
	"fmt"
	"testing"
)

func TestImpPath(t *testing.T) {
	nonVendor := "github.com/foo/bar"
	if nonVendor != impPath(nonVendor) {
		t.Errorf("")
	}
	testCases := []struct {
		input string
		want  string
	}{
		{"foo/bar", "foo/bar"},
		{"vendor/foo/bar", "foo/bar"},
		{"vendor/foo/vendor/bar", "bar"},
		{"/vendor/foo/bar", "foo/bar"},
		{"qux/vendor/foo/bar", "foo/bar"},
		{"qux/vendor/foo/vendor/bar", "bar"},
		{"govendor/foo", "govendor/foo"},
		{"foo/govendor/bar", "foo/govendor/bar"},
		{"vendors/foo", "vendors/foo"},
		{"foo/vendors/bar", "foo/vendors/bar"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input %s", tc.input), func(t *testing.T) {
			if got := impPath(tc.input); got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func Test_typeFromNamedType(t *testing.T) {
	testCases := []struct {
		inputTypePackage string
		inputTypeName    string
		expectedType     *NamedType
	}{
		// not a generic - foo.Bar
		{
			inputTypePackage: "foo",
			inputTypeName:    "Bar",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "Bar",
			},
		},
		// generic - foo.T[int]
		{
			inputTypePackage: "foo",
			inputTypeName:    "T[int]",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "T",
				TypeParams: &TypeParametersType{
					TypeParameters: []Type{
						PredeclaredType("int"),
					},
				},
			},
		},
		// generic - foo.T[int[]]
		{
			inputTypePackage: "foo",
			inputTypeName:    "T[int[]]",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "T",
				TypeParams: &TypeParametersType{
					TypeParameters: []Type{
						PredeclaredType("int[]"),
					},
				},
			},
		},
		// FIXME: broken case
		// generic - foo.T[int[], bool, string[]]
		// {
		// 	inputTypePackage: "foo",
		// 	inputTypeName: "T[int[], bool,  string[]]",
		// 	expectedType: &NamedType{
		// 		Package: "foo",
		// 		Type: "T",
		// 		TypeParams: &TypeParametersType{
		// 			TypeParameters: []Type{
		// 				PredeclaredType("int[]"),
		// 				PredeclaredType("bool"),
		// 				PredeclaredType("string[]"),
		// 			},
		// 		},
		// 	},
		// },
		// generic - foo.T[int, string, int]
		{
			inputTypePackage: "foo",
			inputTypeName:    "T[  int,string,   int]",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "T",
				TypeParams: &TypeParametersType{
					TypeParameters: []Type{
						PredeclaredType("int"),
						PredeclaredType("string"),
						PredeclaredType("int"),
					},
				},
			},
		},
		// generic - foo.T[context.Context]
		{
			inputTypePackage: "foo",
			inputTypeName:    "T[context.Context]",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "T",
				TypeParams: &TypeParametersType{
					TypeParameters: []Type{
						&NamedType{
							Package: "context",
							Type:    "Context",
						},
					},
				},
			},
		},
		// generic - foo.T[context.Context, github.com/foo/bar.X]
		{
			inputTypePackage: "foo",
			inputTypeName:    "T[context.Context , github.com/foo/bar.X ]",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "T",
				TypeParams: &TypeParametersType{
					TypeParameters: []Type{
						&NamedType{
							Package: "context",
							Type:    "Context",
						},
						&NamedType{
							Package: "github.com/foo/bar",
							Type:    "X",
						},
					},
				},
			},
		},
		// generic - foo.T[context.Context, github.com/foo/bar.🤣, int]
		{
			inputTypePackage: "foo",
			inputTypeName:    "T[context.Context , gtihub.com/foo/bar.🤣, int]",
			expectedType: &NamedType{
				Package: "foo",
				Type:    "T",
				TypeParams: &TypeParametersType{
					TypeParameters: []Type{
						&NamedType{
							Package: "context",
							Type:    "Context",
						},
						&NamedType{
							Package: "github.com/foo/bar",
							Type:    "🤣",
						},
						PredeclaredType("int"),
					},
				},
			},
		},
		// FIXME: broken case
		// generic - foo.T[bool[], context.Context[], github.com/foo/bar.X[], int[]]
		// {
		// 	inputTypePackage: "foo",
		// 	inputTypeName:    "foo.T[bool[], context.Context[], github.com/foo/bar.X[], int[]]",
		// 	expectedType: &NamedType{
		// 		Package: "foo",
		// 		Type:    "T",
		// 		TypeParams: &TypeParametersType{
		// 			TypeParameters: []Type{
		// 				PredeclaredType("bool[]"),
		// 				&ArrayType{
		// 					Len: -1,
		// 					Type: &NamedType{
		// 						Package: "context",
		// 						Type:    "Context",
		// 					},
		// 				},
		// 				&ArrayType{
		// 					Len: -1,
		// 					Type: &NamedType{
		// 						Package: "github.com/foo/bar",
		// 						Type:    "X",
		// 					},
		// 				},
		// 				PredeclaredType("int[]"),
		// 			},
		// 		},
		// 	},
		// },
	}

	for idx := range testCases {
		tc := testCases[idx]
		t.Run(fmt.Sprintf("%s.%s", tc.inputTypePackage, tc.inputTypeName), func(t *testing.T) {
			t.Log("input:", tc.inputTypePackage, tc.inputTypeName)

			got := typeFromNamedType(tc.inputTypePackage, tc.inputTypeName)
			gotNamedType, ok := got.(*NamedType)
			if !ok {
				t.Errorf("got %T; want *NamedType", got)
			}
			expected := tc.expectedType
			if gotNamedType.Package != expected.Package {
				t.Errorf("got %s; want %s", gotNamedType.Package, tc.expectedType.Package)
			}
			if expected.TypeParams == nil {
				if gotNamedType.TypeParams != nil {
					t.Errorf("got %s; want nil", gotNamedType.TypeParams)
				}
			} else {
				if gotNamedType.TypeParams == nil {
					t.Errorf("got nil; want %s", expected.TypeParams)
				}

				pm := map[string]string{}
				expectedString := expected.String(pm, "")
				gotString := gotNamedType.String(pm, "")
				if gotString != expectedString {
					t.Errorf("got %q; want %q", gotString, expectedString)
				}
			}
		})
	}
}
