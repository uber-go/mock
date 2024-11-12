package mock

import (
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/mock/mockgen/internal/tests/package_mode"
	"go.uber.org/mock/mockgen/internal/tests/package_mode/cars"
	"go.uber.org/mock/mockgen/internal/tests/package_mode/fuel"
)

// checks, that mocks implement interfaces in compile-time.
// If something breaks, the tests will not be compiled.

var food package_mode.Food = &MockFood{}

var eater package_mode.Eater = &MockEater{}

var animal package_mode.Animal = &MockAnimal{}

var human package_mode.Human = &MockHuman{}

var primate package_mode.Primate = &MockPrimate{}

var car package_mode.Car[fuel.Gasoline] = &MockCar[fuel.Gasoline]{}

var driver package_mode.Driver[fuel.Gasoline, cars.HyundaiSolaris] = &MockDriver[fuel.Gasoline, cars.HyundaiSolaris]{}

var urbanResident package_mode.UrbanResident = &MockUrbanResident{}

var farmer package_mode.Farmer = &MockFarmer{}

func TestInterfaces(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockFarmer(ctrl)
	mock.EXPECT().Breathe()

	farmer := package_mode.Farmer(mock)
	farmer.Breathe()
}
