package mock

import (
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/mock/mockgen/internal/tests/import_mode"
	"go.uber.org/mock/mockgen/internal/tests/import_mode/cars"
	"go.uber.org/mock/mockgen/internal/tests/import_mode/fuel"
)

// checks, that mocks implement interfaces in compile-time.
// If something breaks, the tests will not be compiled.

var food import_mode.Food = &MockFood{}

var eater import_mode.Eater = &MockEater{}

var animal import_mode.Animal = &MockAnimal{}

var human import_mode.Human = &MockHuman{}

var primate import_mode.Primate = &MockPrimate{}

var car import_mode.Car[fuel.Gasoline] = &MockCar[fuel.Gasoline]{}

var driver import_mode.Driver[fuel.Gasoline, cars.HyundaiSolaris] = &MockDriver[fuel.Gasoline, cars.HyundaiSolaris]{}

var urbanResident import_mode.UrbanResident = &MockUrbanResident{}

var farmer import_mode.Farmer = &MockFarmer{}

func TestInterfaces(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockFarmer(ctrl)
	mock.EXPECT().Breathe()

	farmer := import_mode.Farmer(mock)
	farmer.Breathe()
}
