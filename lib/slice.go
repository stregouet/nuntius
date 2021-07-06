package lib

import "sort"

func IsSliceIntEqual(a, b []int) bool {
    if len(a) != len(b) {
        return false
    }
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}

// test if a and b are equal, whatever the order
// note that this function will sort a and b inplace
func IsCountEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
    sort.Strings(a)
    sort.Strings(b)
	for i, el := range a {
		if el != b[i] {
			return false
		}
	}
	return true
}
