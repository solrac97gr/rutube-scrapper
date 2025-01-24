package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

type Influencer struct {
	Name      string
	Followers string
	Index     int // Track the original order
}

func main() {
	// Open the file containing the list of Rutube links
	file, err := os.Open("influencers.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		links = append(links, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Scrape the links concurrently while preserving order
	influencers := scrapeLinksConcurrently(links)

	// Sort influencers by their original index to maintain order
	sort.Slice(influencers, func(i, j int) bool {
		return influencers[i].Index < influencers[j].Index
	})

	// Save the results to an Excel file
	err = saveToExcel("influencers.xlsx", influencers)
	if err != nil {
		log.Fatalf("Failed to save to Excel: %v", err)
	}

	fmt.Println("Scraping complete! Results saved to 'influencers.xlsx'")
}

func scrapeRutubeProfile(url string, index int) (Influencer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Influencer{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Influencer{}, fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Influencer{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract the name from the title attribute of the <h1> tag
	name := doc.Find("h1.wdp-feed-banner-module__wdp-feed-banner__title-text").AttrOr("title", "")
	name = strings.TrimSpace(name)

	// Extract followers and clean the string
	followers := doc.Find(".wdp-feed-banner-module__wdp-feed-banner__title p").Text()
	followers = strings.TrimSpace(followers)

	// Clean up the followers string to remove "подписчиков" and keep only the number
	re := regexp.MustCompile(`[^\d]`)
	followers = re.ReplaceAllString(followers, "")

	if name == "" || followers == "" {
		return Influencer{}, fmt.Errorf("failed to extract name or followers, check selectors")
	}

	return Influencer{Name: name, Followers: followers, Index: index}, nil
}

func scrapeLinksConcurrently(links []string) []Influencer {
	var wg sync.WaitGroup
	influencers := make([]Influencer, len(links)) // Pre-allocate slice to preserve order
	mu := sync.Mutex{}

	for i, link := range links {
		wg.Add(1)
		go func(i int, link string) {
			defer wg.Done()
			influencer, err := scrapeRutubeProfile(link, i)
			if err != nil {
				log.Printf("Error scraping %s: %v", link, err)
				return
			}
			mu.Lock()
			influencers[i] = influencer // Store in the correct index
			mu.Unlock()
		}(i, link)
	}
	wg.Wait()
	return influencers
}

func saveToExcel(filename string, influencers []Influencer) error {
	f := excelize.NewFile()

	// Create a sheet and set headers
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Name")
	f.SetCellValue(sheet, "B1", "Followers")

	// Populate the sheet with data
	for i, inf := range influencers {
		row := i + 2 // Start from the second row
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), inf.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), inf.Followers)
	}

	// Save the Excel file
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
