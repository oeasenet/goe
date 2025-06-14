package utils

func Convert[T any](value string, convertor func(string) (T, error), defaultValue ...T) (*T, error) {
	result, err := convertor(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return &defaultValue[0], nil
		}
		return nil, err
	}
	return &result, nil
}
