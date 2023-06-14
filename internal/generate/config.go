package generate

type NewConfigFunc func() any

func newConfig[T any]() NewConfigFunc {
	return func() any {
		t := new(T)
		return t
	}
}
