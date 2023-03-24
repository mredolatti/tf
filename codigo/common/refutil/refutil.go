package refutil

func Ref[T any](t T) *T {
	return &t
}
