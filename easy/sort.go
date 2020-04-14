package easy

import (
	"sort"
)

func SearchSortedInts(slice []int, elem int) int {
	return sort.Search(len(slice), func(i int) bool { return slice[i] >= elem })
}

func SearchSortedInt32s(slice []int32, elem int32) int {
	return sort.Search(len(slice), func(i int) bool { return slice[i] >= elem })
}

func SearchSortedInt64s(slice []int64, elem int64) int {
	return sort.Search(len(slice), func(i int) bool { return slice[i] >= elem })
}

func SearchSortedStrings(slice []string, elem string) int {
	return sort.Search(len(slice), func(i int) bool { return slice[i] >= elem })
}

func InSortedInts(slice []int, elem int) bool {
	length := len(slice)
	if length == 0 {
		return false
	}
	if length == 1 || slice[0] == slice[length-1] {
		return slice[0] == elem
	}

	var less func(i int) bool
	if slice[0] <= slice[length-1] {
		// ascending order
		less = func(i int) bool { return slice[i] >= elem }
	} else {
		// descending order
		less = func(i int) bool { return slice[i] <= elem }
	}
	i := sort.Search(length, less)
	return i < len(slice) && slice[i] == elem
}

func InSortedInt32s(slice []int32, elem int32) bool {
	length := len(slice)
	if length == 0 {
		return false
	}
	if length == 1 || slice[0] == slice[length-1] {
		return slice[0] == elem
	}

	var less func(i int) bool
	if slice[0] <= slice[length-1] {
		// ascending order
		less = func(i int) bool { return slice[i] >= elem }
	} else {
		// descending order
		less = func(i int) bool { return slice[i] <= elem }
	}
	i := sort.Search(length, less)
	return i < len(slice) && slice[i] == elem
}

func InSortedInt64s(slice []int64, elem int64) bool {
	length := len(slice)
	if length == 0 {
		return false
	}
	if length == 1 || slice[0] == slice[length-1] {
		return slice[0] == elem
	}

	var less func(i int) bool
	if slice[0] <= slice[length-1] {
		// ascending order
		less = func(i int) bool { return slice[i] >= elem }
	} else {
		// descending order
		less = func(i int) bool { return slice[i] <= elem }
	}
	i := sort.Search(length, less)
	return i < len(slice) && slice[i] == elem
}

func InSortedStrings(slice []string, elem string) bool {
	length := len(slice)
	if length == 0 {
		return false
	}
	if length == 1 || slice[0] == slice[length-1] {
		return slice[0] == elem
	}

	var less func(i int) bool
	if slice[0] <= slice[length-1] {
		// ascending order
		less = func(i int) bool { return slice[i] >= elem }
	} else {
		// descending order
		less = func(i int) bool { return slice[i] <= elem }
	}
	i := sort.Search(length, less)
	return i < len(slice) && slice[i] == elem
}
