package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Index returns the first index of the target string `t`, or
// -1 if no match is found.
func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// Include returns `true` if the target string t is in the
// slice.
func Include(vs []string, t string) bool {
	return Index(vs, t) >= 0
}

// Any returns `true` if one of the strings in the slice
// satisfies the predicate `f`.
func Any(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if f(v) {
			return true
		}
	}
	return false
}

// All returns `true` if every strings in the slice
// satisfy the predicate `f`.
func All(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing all strings in the
// slice that satisfy the predicate `f`.
func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Map returns a new slice containing the results of applying
// the function `f` to each string in the original slice.
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// Create a unique list of values (remove duplicates)
func Uniq(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func Equal(a, b *[]string) bool {
	if a == b && a == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	left := *a
	right := *b

	if len(left) != len(right) {
		return false
	}
	for _, v := range right {
		if !Include(left, v) {
			return false
		}
	}
	return true
}

func MapEquals(a, b *map[string][]string) bool {
	if a == b && a == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	left := *a
	right := *b

	for key, value := range right {
		if otherValue, exist := left[key]; exist {
			if !Equal(&otherValue, &value) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func MapConcat(a, b *map[string][]string) *map[string][]string {
	if a == b && a == nil {
		return nil
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	left := *a
	right := *b

	result := map[string][]string{}
	for key, _ := range right {
		if _, exist := left[key]; exist {
			result[key] = Uniq(append(left[key], right[key]...))
		} else {
			result[key] = Uniq(right[key])
		}
	}
	for key, _ := range left {
		if _, exist := right[key]; !exist {
			result[key] = Uniq(left[key])
		}
	}
	return &result
}

func ExtractLDAPCN(DN string) (string, error) {
	CN := regexp.MustCompile(`CN=([^,]+)`)
	cn := CN.FindStringSubmatch(DN)

	if len(cn) < 1 {
		return "", errors.New(fmt.Sprintf("LDAP CN cannot be extracted from the DN: %s", DN))
	}

	return strings.ToLower(cn[1]), nil
}

func ExtractProjectKey(tenant string) string {
	if len(tenant) < 4 {
		return strings.ToLower(tenant)
	} else {

		return strings.ToLower(tenant[0:2] + tenant[len(tenant)-2:])
	}
}
