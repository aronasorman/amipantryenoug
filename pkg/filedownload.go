package pkg

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Choose files to download from the old and new pantry site.
// It will pick one file from each file type defined in FileTypesToDeepCheck.
func (s *Scraper) ChooseFiles() map[string]string {
	// just pick the first one we see for each file type lol

	res := map[string]string{}

	for _, url := range s.fileLinks {
		for _, fileExt := range FileTypesToDeepCheck {
			if strings.HasSuffix(url, fileExt) {
				_, ok := res[fileExt]
				if !ok {
					res[fileExt] = url
				}
			}
		}
	}

	return res
}

// returns a map of all files checked and whether they match.
func (s *Scraper) DeepCheckFiles(newHost string) map[string]bool {
	files := s.ChooseFiles()
	res := map[string]bool{}

	for _, oldUrl := range files {
		newUrl := newHostUrl(oldUrl, newHost)
		fmt.Printf("Checking %s\n", newUrl)
		res[oldUrl] = isSameChecksum(oldUrl, newUrl)
	}

	return res
}

// getFileChecksum returns the md5 checksum of the file at the given url.
func getFileChecksum(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	hasher := md5.New()
	tr := io.TeeReader(resp.Body, hasher)

	if _, err := io.Copy(io.Discard, tr); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// isSameChecksum accepts an old pantry url and checks the same file
// in the new pantry url. Returns true if the checksums match. False
// if they don't match.
func isSameChecksum(oldUrl, newUrl string) bool {
	oldChecksum, err := getFileChecksum(oldUrl)
	if err != nil {
		return false
	}

	newChecksum, err := getFileChecksum(newUrl)
	if err != nil {
		return false
	}

	return oldChecksum == newChecksum
}
