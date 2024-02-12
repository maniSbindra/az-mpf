package domain

import "sort"

func getMapWithUniqueValues(m map[string][]string) map[string][]string {
	sm := make(map[string][]string)
	for key, vals := range m {
		vals = getUniqueSlice(vals)
		sort.Strings(vals)
		sm[key] = vals
	}
	return sm
}

func getUniqueSlice(s []string) []string {
	uniqueSlice := make([]string, 0, len(s))
	m := make(map[string]bool)
	for _, val := range s {
		if _, ok := m[val]; !ok {
			m[val] = true
			uniqueSlice = append(uniqueSlice, val)
		}
	}
	return uniqueSlice
}
