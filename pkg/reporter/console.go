package reporter

import (
	"dupe-file-checker/pkg/scanner"
	"fmt"
)

func PrintDuplicates(groups []scanner.DuplicateGroup) {
	if len(groups) == 0 {
		fmt.Println("No duplicates found")
		return
	}

	fmt.Printf("Found %d duplicate groups:\n\n", len(groups))

	groupNum := 1
	for _, group := range groups {
		if len(group.Files) < 2 {
			continue
		}

		fmt.Printf("Group %d (%d duplicates):\n", groupNum, len(group.Files))
		for _, path := range group.Files {
			fmt.Printf("  - %s\n", path)
		}
		fmt.Println()
		groupNum++
	}
}
