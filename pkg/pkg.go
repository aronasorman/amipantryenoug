package pkg

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	colly "github.com/gocolly/colly/v2"
)

const ScraperUserAgent = "amipantryenoughscraper/0.1"

var FileTypesToDownload = []string{
	"pdf", "deb", "exe", "zip", "pex", "torrent", "img",
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
			for _, fileType := range FileTypesToDownload {
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
	for _, link := range s.fileLinks {
		err := checkAgainstNewUrl(link, host)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			notFound++
		}
	}

	return notFound
}

// check that the file in the old url is present in the new host
// raises an error if the url is not present, or for any url parsing/fetching
// errors. Otherwise, returns nil.
func checkAgainstNewUrl(oldUrl, newHost string) error {
	oldUri, err := url.Parse(oldUrl)
	if err != nil {
		return err
	}

	newUrl := fmt.Sprintf("%s://%s", oldUri.Scheme, path.Join(newHost, oldUri.Path))
	resp, err := http.Head(newUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", newUrl, resp.Status)
	}

	return nil
}
