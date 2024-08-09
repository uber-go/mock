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
				ifaces:      []string{"ImmortalHelldiver"},
			},
			expectedErr: "failed to extract interfaces from package: interface ImmortalHelldiver does not exists",
		},
		{
			name: "error: search for struct instead of interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Enemy"},
			},
			expectedErr: "failed to extract interfaces from package: failed to parse interface: " +
				"Enemy is not an interface. it is a struct{Name string; Fraction string; Hp int}",
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
				ifaces:      []string{"DemocracyFan"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "DemocracyFan",
						Methods: []*model.Method{
							{Name: "ILoveDemocracy"},
							{Name: "YouWillNeverDestroyOurWayOfLife"},
						},
					},
				},
			},
		},
		{
			name: "success: interface with generic",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Shooter"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Shooter",
						Methods: []*model.Method{
							{
								Name: "Gun",
								Out: []*model.Parameter{
									{Type: &model.NamedType{Type: "GunType"}},
								},
							},
							{
								Name: "Reload",
								Out: []*model.Parameter{
									{Type: model.PredeclaredType("bool")},
								},
							},
							{
								Name: "Shoot",
								In: []*model.Parameter{
									{Name: "times", Type: model.PredeclaredType("int")},
								},
								Out: []*model.Parameter{
									{Type: model.PredeclaredType("bool")},
									{Type: &model.NamedType{Type: "error"}},
								},
								Variadic: &model.Parameter{
									Name: "targets",
									Type: &model.PointerType{
										Type: &model.NamedType{
											Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
											Type:    "Enemy",
										},
									},
								},
							},
						},
						TypeParams: []*model.Parameter{
							{
								Name: "ProjectileType",
								Type: &model.NamedType{
									Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
									Type:    "Projectile",
								},
							},
							{
								Name: "GunType",
								Type: &model.NamedType{
									Package: "go.uber.org/mock/mockgen/internal/tests/import_mode",
									Type:    "Gun",
									TypeParams: &model.TypeParametersType{
										TypeParameters: []model.Type{
											&model.NamedType{
												Type: "ProjectileType",
											},
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
			name: "success: interface with embedded interfaces",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"Helldiver"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "Helldiver",
						Methods: []*model.Method{
							{
								Name: "AvailableStratagems",
								Out: []*model.Parameter{
									{
										Type: &model.ArrayType{
											Len: -1,
											Type: &model.NamedType{
												Package: "go.uber.org/mock/mockgen/internal/tests/import_mode/stratagems",
												Type:    "Stratagem",
											},
										},
									},
								},
							},
							{Name: "ILoveDemocracy"},
							{Name: "YouWillNeverDestroyOurWayOfLife"},
						},
					},
				},
			},
		},
		{
			name: "success: alias to interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"SuperEarthCitizen"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "SuperEarthCitizen",
						Methods: []*model.Method{
							{Name: "ILoveDemocracy"},
							{Name: "YouWillNeverDestroyOurWayOfLife"},
						},
					},
				},
			},
		},
		{
			name: "success: embedded anonymous interface",
			args: args{
				packageName: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				ifaces:      []string{"AgitationCampaign"},
			},
			expected: &model.Package{
				Name:    "import_mode",
				PkgPath: "go.uber.org/mock/mockgen/internal/tests/import_mode",
				Interfaces: []*model.Interface{
					{
						Name: "AgitationCampaign",
						Methods: []*model.Method{
							{Name: "BecomeAHelldiver"},
							{Name: "BecomeAHero"},
							{Name: "BecomeALegend"},
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
