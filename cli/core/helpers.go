package core

import "fmt"

func ToStringList[T any](values []T) []string {
	var stringValues = make([]string, len(values))
	for i, v := range values {
		stringValues[i] = fmt.Sprintf("%v", v)
	}

	return stringValues
}
