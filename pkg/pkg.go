package pkg

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	colly "github.com/gocolly/colly/v2"
)

const ScraperUserAgent = "amipantryenoughscraper/0.1"

// FileTypesToCheck is a list of file types to check for existence in pantry.
var FileTypesToCheck = []string{
	"pdf", "deb", "exe", "zip", "pex", "torrent", "img",
}

// FileTypesToDeepCheck is a list of file types to download and check for hash
// mismatches.
var FileTypesToDeepCheck = []string{
	"pdf", "deb", "pex", "zip", "torrent",
}

type Scraper struct {
	c         *colly.Collector
	fileLinks []string
}

func NewScraper() *Scraper {
	c := colly.NewCollector(
		colly.AllowedDomains("pantry.learningequality.org", "pantry-new.learningequality.org"),
		colly.UserAgent(ScraperUserAgent),
		colly.Async(true),
	)
	s := &Scraper{
		c:         c,
		fileLinks: []string{},
	}
	s.setupCollector()

	return s
}

func (s *Scraper) FetchAllFileLinks(startingUrl string) ([]string, error) {
	err := s.c.Visit(startingUrl)
	if err != nil {
		return nil, err
	}
	s.c.Wait()

	return s.fileLinks, nil
}

func (s *Scraper) setupCollector() {
	s.c.OnHTML("a", func(e *colly.HTMLElement) {
		if e.Text != "../" { // ignore "../" links
			// if our link ends in one of the filetypes, record it for comparison later
			link := e.Request.AbsoluteURL(e.Attr("href"))
			for _, fileType := range FileTypesToCheck {
				if strings.HasSuffix(link, fileType) {
					fmt.Printf("Found file: %s\n", link)
					s.fileLinks = append(s.fileLinks, link)
					return
				}
			}

			// did not early return, recurse on this link
			childLink := e.Request.AbsoluteURL(e.Attr("href"))
			s.c.Visit(childLink)
		}
	})
}

// Check the collected file links against the new host.
// Returns the number of files that were not found on the new host.
func (s *Scraper) VerifyNewHost(host string) int {
	var notFound int
	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, link := range s.fileLinks {
		go checkAgainstNewUrl(link, host, &wg, errChan)
		wg.Add(1)
	}

	go func() {
		for err := range errChan {
			fmt.Printf("Error: %s\n", err)
			notFound++
		}
	}()

	wg.Wait()
	close(errChan)
	return notFound
}

// check that the file in the old url is present in the new host
// raises an error if the url is not present, or for any url parsing/fetching
// errors. Otherwise, returns nil.
func checkAgainstNewUrl(oldUrl, newHost string, wg *sync.WaitGroup, errChan chan<- error) {
	newUrl := newHostUrl(oldUrl, newHost)
	fmt.Printf("checking %s\n", newUrl)
	resp, err := http.Head(newUrl)
	if err != nil {
		errChan <- err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s: %s", newUrl, resp.Status)
	}

	wg.Done()
}

func newHostUrl(oldUrl, newHost string) string {
	oldUri, _ := url.Parse(oldUrl)
	return fmt.Sprintf("%s://%s", oldUri.Scheme, path.Join(newHost, oldUri.Path))
}
