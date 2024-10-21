package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/mockgen/model"
)

func Test_packageModeParser_parsePackage(t *testing.T) {
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
				packageName: "foo/bar/another_package",
				ifaces:      []string{"Human"},
			},
			expectedErr: "load package: -: package foo/bar/another_package is not in std",
		},
		{
			name: "error: interface does not exist",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Alien"},
			},
			expectedErr: "extract interfaces from package: interface Alien does not exist",
		},
		{
			name: "error: search for struct instead of interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Work"},
			},
			expectedErr: "extract interfaces from package: parse interface: " +
				"error parsing Work: " +
				"Work is not an interface. it is a *types.Struct",
		},
		{
			name: "error: search for constraint instead of interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Counter"},
			},
			expectedErr: "extract interfaces from package: parse interface: " +
				"error parsing Counter: " +
				"interface Counter is a constraint",
		},
		{
			name: "success: simple interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Food"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Eater"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Eater",
						Methods: []*model.Method{
							{
								Name: "Eat",
								Variadic: &model.Parameter{
									Name: "foods",
									Type: &model.NamedType{
										Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Car"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
											Package: "go.uber.org/mock/mockgen/internal/tests/package_mode/cars",
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
									Package: "go.uber.org/mock/mockgen/internal/tests/package_mode/fuel",
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
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Animal"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
										Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Primate"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
										Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Human"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
										Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
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
			name: "success: interfaces with aliases in params and returns",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				ifaces:      []string{"Earth"},
			},
			expected: &model.Package{
				Name:    "package_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/package_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Earth",
						Methods: []*model.Method{
							{
								Name: "AddHumans",
								In: []*model.Parameter{
									{
										Type: model.PredeclaredType("int"),
									},
								},
								Out: []*model.Parameter{
									{
										Type: &model.ArrayType{
											Len: -1, // slice
											Type: &model.NamedType{
												Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
												Type:    "Primate",
											},
										},
									},
								},
							},
							{
								Name: "HumanPopulation",
								Out: []*model.Parameter{
									{
										Type: model.PredeclaredType("int"),
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
			parser := packageModeParser{}
			actual, err := parser.parsePackage(tt.args.packageName, tt.args.ifaces)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}
