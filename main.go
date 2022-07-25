package main

import (
	"fmt"

	"github.com/aronasorman/amipantryenough/pkg"
)

func main() {
	scraper := pkg.NewScraper()
	_, err := scraper.FetchAllFileLinks("https://pantry.learningequality.org/downloads")
	if err != nil {
		panic(err)
	}

	filesNotFound := scraper.VerifyNewHost("pantry-new.learningequality.org")
	fmt.Printf("\n\n\n%d files not found on new host\n", filesNotFound)

	fmt.Printf("\n\nDeep checking files...\n")
	deepCheckedFiles := scraper.DeepCheckFiles("pantry-new.learningequality.org")
	filesNotMatched := 0
	for file, match := range deepCheckedFiles {
		if !match {
			fmt.Printf("%s not matching on new host\n", file)
			filesNotMatched++
		}
	}

	fmt.Printf("\n\n%d files not matching on new host\n", filesNotMatched)
}
