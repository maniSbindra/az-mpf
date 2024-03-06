package presentation

import (
	"fmt"
	"io"
	"sort"
)

func (d *displayConfig) displayText(w io.Writer) error {

	sm := d.result.RequiredPermissions

	defaultPerms := sm[d.displayOptions.DefaultResourceGroupResourceID]

	// sort permissions
	// defaultPerms = getUniqueSlice(defaultPerms)
	sort.Strings(defaultPerms)

	// print permissions for default scope
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("Permissions Required:")
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	for _, perm := range defaultPerms {
		fmt.Println(perm)
	}
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println()

	if !d.displayOptions.ShowDetailedOutput {
		return nil
	}

	fmt.Println()
	fmt.Println("Break down of permissions by different resource types:")
	fmt.Println()

	// print permissions for other scopes
	for scope, perms := range sm {
		if scope == d.displayOptions.DefaultResourceGroupResourceID {
			continue
		}

		// perms = getUniqueSlice(perms)
		sort.Strings(perms)

		fmt.Printf("Permissions required for %s: \n", scope)
		for _, perm := range perms {
			fmt.Printf("%s\n", perm)
		}
		fmt.Println("--------------")
		fmt.Println()
		fmt.Println()
	}
	return nil
}
