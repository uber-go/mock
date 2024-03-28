package typed_inorder

import (
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestInteract(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockAnimal := NewMockAnimal(ctrl)
	gomock.InOrder(
		mockAnimal.EXPECT().Feed("burguir").DoAndReturn(func(s string) error {
			if s != "chocolate" {
				return nil
			}
			return fmt.Errorf("Dogs can't eat chocolate!")
		}),
		mockAnimal.EXPECT().GetSound().Return("Woof!"),
	)
	_, err := Interact(mockAnimal, "burguir")
	if err != nil {
		t.Fatalf("sad")
	}
}
