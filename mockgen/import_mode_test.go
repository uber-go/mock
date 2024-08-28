package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/mockgen/model"
)

func Test_importModeParser_parsePackage(t *testing.T) {
	type args struct {
		packageName string
		ifaces      []string
	}
	tests := []struct {
		name        string
		args        args
		expected    *model.Package
		expectedErr string
	}{
		{
			name: "error: no found package",
			args: args{
				packageName: "",
				ifaces:      []string{"ImmortalHelldiver"},
			},
			expectedErr: "failed to load package: package  not found",
		},
		{
			name: "error: interface does not exists",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Alien"},
			},
			expectedErr: "failed to extract interfaces from package: interface Alien does not exists",
		},
		{
			name: "error: search for struct instead of interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Work"},
			},
			expectedErr: "failed to extract interfaces from package: failed to parse interface: " +
				"Work is not an interface. it is a struct{Name string}",
		},
		{
			name: "error: search for constraint instead of interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Counter"},
			},
			expectedErr: "failed to extract interfaces from package: failed to parse interface: " +
				"interface Counter is a constraint",
		},
		{
			name: "success: simple interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Food"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Food",
						Methods: []*model.Method{
							{
								Name: "Calories",
								Out: []*model.Parameter{
									{Type: model.PredeclaredType("int")},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success: interface with variadic args",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Eater"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Eater",
						Methods: []*model.Method{
							{
								Name: "Eat",
								Variadic: &model.Parameter{
									Name: "foods",
									Type: &model.NamedType{
										Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
										Type:    "Food",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success: interface with generic",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Car"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Car",
						Methods: []*model.Method{
							{
								Name: "Brand",
								Out: []*model.Parameter{
									{Type: model.PredeclaredType("string")},
								},
							},
							{
								Name: "FuelTank",
								Out: []*model.Parameter{
									{
										Type: &model.NamedType{
											Package: "go.uber.org/mock/mockgen/internal/tests/import_mode/cars",
											Type:    "FuelTank",
											TypeParams: &model.TypeParametersType{
												TypeParameters: []model.Type{
													&model.NamedType{
														Type: "FuelType",
													},
												},
											},
										},
									},
								},
							},
							{
								Name: "Refuel",
								In: []*model.Parameter{
									{
										Name: "fuel",
										Type: &model.NamedType{Type: "FuelType"},
									},
									{
										Name: "volume",
										Type: model.PredeclaredType("int"),
									},
								},
								Out: []*model.Parameter{
									{Type: &model.NamedType{Type: "error"}},
								},
							},
						},
						TypeParams: []*model.Parameter{
							{
								Name: "FuelType",
								Type: &model.NamedType{
									Package: "go.uber.org/mock/mockgen/internal/tests/import_mode/fuel",
									Type:    "Fuel",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success: interface with embedded interfaces",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Animal"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Animal",
						Methods: []*model.Method{
							{Name: "Breathe"},
							{
								Name: "Eat",
								Variadic: &model.Parameter{
									Name: "foods",
									Type: &model.NamedType{
										Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
										Type:    "Food",
									},
								},
							},
							{
								Name: "Sleep",
								In: []*model.Parameter{
									{
										Name: "duration",
										Type: &model.NamedType{
											Package: "time",
											Type:    "Duration",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success: subtype of interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Primate"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Primate",
						Methods: []*model.Method{
							{Name: "Breathe"},
							{
								Name: "Eat",
								Variadic: &model.Parameter{
									Name: "foods",
									Type: &model.NamedType{
										Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
										Type:    "Food",
									},
								},
							},
							{
								Name: "Sleep",
								In: []*model.Parameter{
									{
										Name: "duration",
										Type: &model.NamedType{
											Package: "time",
											Type:    "Duration",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success: alias to interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Human"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Human",
						Methods: []*model.Method{
							{Name: "Breathe"},
							{
								Name: "Eat",
								Variadic: &model.Parameter{
									Name: "foods",
									Type: &model.NamedType{
										Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
										Type:    "Food",
									},
								},
							},
							{
								Name: "Sleep",
								In: []*model.Parameter{
									{
										Name: "duration",
										Type: &model.NamedType{
											Package: "time",
											Type:    "Duration",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := importModeParser{}
			actual, err := parser.parsePackage(tt.args.packageName, tt.args.ifaces)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}
