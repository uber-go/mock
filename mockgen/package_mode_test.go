package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
										Type: &model.NamedType{
											Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
											Type: "HumansCount",
										},
									},
								},
								Out: []*model.Parameter{
									{
										Type: &model.ArrayType{
											Len: -1, // slice
											Type: &model.NamedType{
												Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
												Type:    "Human",
											},
										},
									},
								},
							},
							{
								Name: "HumanPopulation",
								Out: []*model.Parameter{
									{
										Type: &model.NamedType{
											Package: "go.uber.org/mock/mockgen/internal/tests/package_mode",
											Type: "HumansCount",
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

// This tests the alias replacement behavior of package mode.
// TODO(joaks): Update this once we remove the replacement logic
// when we bump go.mod to 1.23.
func TestAliases(t *testing.T) {
	packageName := "go.uber.org/mock/mockgen/internal/tests/alias"
	for _, tt := range []struct {
		desc     string
		iface    string
		expected *model.Interface
	}{
		{
			desc:  "interface with alias references elsewhere",
			iface: "Fooer",
			expected: &model.Interface{
				Name: "Fooer",
				Methods: []*model.Method{{
					Name: "Foo",
				}},
			},
		},
		{
			desc:  "alias to an interface in the same package",
			iface: "FooerAlias",
			expected: &model.Interface{
				Name: "FooerAlias",
				Methods: []*model.Method{{
					Name: "Foo",
				}},
			},
		},
		{
			desc:  "interface that takes/returns aliases from same package",
			iface: "Barer",
			expected: &model.Interface{
				Name: "Barer",
				Methods: []*model.Method{{
					Name: "Bar",
					In: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "FooerAlias",
						},
					}},
					Out: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "FooerAlias",
						},
					}},
				}},
			},
		},
		{
			desc:  "alias to an interface that takes/returns aliases from same package",
			iface: "BarerAlias",
			expected: &model.Interface{
				Name: "BarerAlias",
				Methods: []*model.Method{{
					Name: "Bar",
					In: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "FooerAlias",
						},
					}},
					Out: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "FooerAlias",
						},
					}},
				}},
			},
		},
		{
			desc:  "interface that refers to underlying name when alias is referenced in package",
			iface: "Bazer",
			expected: &model.Interface{
				Name: "Bazer",
				Methods: []*model.Method{{
					Name: "Baz",
					In: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "Fooer",
						},
					}},
					Out: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "Fooer",
						},
					}},
				}},
			},
		},
		{
			desc:  "interface that refers to an alias to another package type",
			iface: "QuxerConsumer",
			expected: &model.Interface{
				Name: "QuxerConsumer",
				Methods: []*model.Method{{
					Name: "Consume",
					In: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "QuxerAlias",
						},
					}},
					Out: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias",
							Type:    "QuxerAlias",
						},
					}},
				}},
			},
		},
		{
			desc:  "interface that refers to another package alias to another package type",
			iface: "QuuxerConsumer",
			expected: &model.Interface{
				Name: "QuuxerConsumer",
				Methods: []*model.Method{{
					Name: "Consume",
					In: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias/subpkg",
							Type:    "Quuxer",
						},
					}},
					Out: []*model.Parameter{{
						Type: &model.NamedType{
							Package: "go.uber.org/mock/mockgen/internal/tests/alias/subpkg",
							Type:    "Quuxer",
						},
					}},
				}},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			var parser packageModeParser
			actual, err := parser.parsePackage(packageName, []string{tt.iface})
			require.NoError(t, err)
			require.NotNil(t, actual)
			require.Len(t, actual.Interfaces, 1)
			assert.Equal(t, "alias", actual.Name)
			assert.Equal(t, "go.uber.org/mock/mockgen/internal/tests/alias", actual.PkgPath)
			assert.Equal(t, tt.expected, actual.Interfaces[0])
		})
	}
}
