package inorder


//go:generate mockgen -package inorder -source=inorder.go -destination=mock.go -typed
type Animal interface {
        GetNoise() string
        Feed(string) error
}

func Interact(a Animal, food string) (string, error) {
        if err := a.Feed(food); err != nil {
                return "", err
        }
        return a.GetNoise(), nil
}
