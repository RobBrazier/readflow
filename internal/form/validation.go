package form

import "errors"

func ValidationMinValues[T comparable](min int) func([]T) error {
	return func(t []T) error {
		if len(t) < min {
			return errors.New("You must select at least one")
		}
		return nil
	}
}

func ValidationRequired[T comparable]() func(T) error {
	return func(t T) error {
		var empty T
		if t == empty {
			return errors.New("This field is required")
		}
		return nil
	}
}
