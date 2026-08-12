package resolver

func validateNumber[T int | int64 | float64](value any) (T, bool) {
	v, ok := value.(float64)
	if !ok {
		var zero T
		return zero, false
	}
	return T(v), true
}

func validateString(value any) (string, bool) {
	v, ok := value.(string)
	if len(v) == 0 {
		return v, false
	}
	return v, ok
}

func validateDataTypeSlice[T any](value any) ([]T, bool) {
	typedValues, ok := value.([]any)
	if !ok {
		return nil, false
	}

	values := make([]T, 0, len(typedValues))
	for _, v := range typedValues {
		value, ok := v.(T)
		if ok {
			values = append(values, value)
		}
	}
	return values, len(values) != 0
}

func validateMap(value any) (map[string]any, bool) {
	v, ok := value.(map[string]any)
	return v, ok
}
