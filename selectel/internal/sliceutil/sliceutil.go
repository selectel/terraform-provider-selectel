package sliceutil

import "slices"

func Union[T comparable](base, extra []T) []T {
	seen := make(map[T]struct{}, len(base))
	for _, v := range base {
		seen[v] = struct{}{}
	}

	result := slices.Clone(base)
	for _, v := range extra {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

func Subtract[T comparable](base, removed []T) []T {
	drop := make(map[T]struct{}, len(removed))
	for _, v := range removed {
		drop[v] = struct{}{}
	}

	result := make([]T, 0, len(base))
	for _, v := range base {
		if _, ok := drop[v]; !ok {
			result = append(result, v)
		}
	}

	return result
}

func Intersect[T comparable](base, other []T) []T {
	keep := make(map[T]struct{}, len(other))
	for _, v := range other {
		keep[v] = struct{}{}
	}

	result := make([]T, 0, len(base))
	for _, v := range base {
		if _, ok := keep[v]; ok {
			result = append(result, v)
		}
	}

	return result
}
