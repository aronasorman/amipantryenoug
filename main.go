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
	fmt.Printf("%d files not found on new host\n", filesNotFound)
}
