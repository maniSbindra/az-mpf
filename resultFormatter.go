package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
)

func getMapWithUniquePermissionsForEachScope(m map[string][]string) map[string][]string {
	sm := make(map[string][]string)
	for scope, perms := range m {
		perms = getUniqueSlice(perms)
		sort.Strings(perms)
		sm[scope] = perms
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

func (m *MinPermFinder) FormatResultAsJSON() {
	sortedUniqueScopePermissionMap := getMapWithUniquePermissionsForEachScope(m.scopePermissionMap)
	jsonBytes, err := json.Marshal(sortedUniqueScopePermissionMap)
	if err != nil {
		log.Fatalf("Error converting output to JSON :%v \n", err)
	}
	fmt.Println(string(jsonBytes))

	// print permissions for default scope in JSON format by serelizing m.scopePermissionMap
}

func (m *MinPermFinder) FormatResult() {

	if m.JSONOutput {
		m.FormatResultAsJSON()
		return
	}
	sm := m.scopePermissionMap

	defaultPerms := sm[m.ResourceGroupResourceID]

	// sort permissions
	defaultPerms = getUniqueSlice(defaultPerms)
	sort.Strings(defaultPerms)

	// print permissions for default scope
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("Permissions Assigned to Service Principal for Resource Group: ", m.ResourceGroupResourceID)
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	for _, perm := range defaultPerms {
		fmt.Println(perm)
	}
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println()

	if !m.ShowDetailedOutput {
		return
	}

	fmt.Println()
	fmt.Println("Break down of permissions by different resource types:")
	fmt.Println()

	// print permissions for other scopes
	for scope, perms := range sm {
		if scope == m.ResourceGroupResourceID {
			continue
		}

		perms = getUniqueSlice(perms)
		sort.Strings(perms)

		// create map to get unique permissions from perms
		// permMap := make(map[string]bool)
		// for _, perm := range perms {
		// 	permMap[perm] = true
		// }
		// print permissions for scope

		// for perm := range permMap {
		// 	fmt.Println(perm)
		// }
		fmt.Printf("Permissions required for %s: \n", scope)
		for _, perm := range perms {
			fmt.Printf("%s\n", perm)
		}
		fmt.Println("--------------")
		fmt.Println()
		fmt.Println()

	}

}
