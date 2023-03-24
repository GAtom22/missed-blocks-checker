package utils

func StringInSlice(first string, list []string) bool {
	for _, second := range list {
		if first == second {
			return true
		}
	}
	return false
}

func RemoveFromSlice(slice []string, r string) []string {
	for i, v := range slice {
		if v == r {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func FilterMap[T any](source map[string]T, f func(T) bool) map[string]T {
	n := make(map[string]T, len(source))
	for key, value := range source {
		if f(value) {
			n[key] = value
		}
	}
	return n
}

func MapToSlice[T any](source map[string]T) []T {
	n := make([]T, len(source))

	index := 0
	for _, value := range source {
		n[index] = value
		index++
	}

	return n
}
