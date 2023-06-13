package mutate

type NewConfigFunc func() any

func newAnyMap() any {
	return make(map[string]any)
}

func newStringMap() any {
	return make(map[string]string)
}

func newAnySlice() any {
	return make([]any, 0)
}

func newConfig[T any]() NewConfigFunc {
	return func() any {
		return new(T)
	}
}
