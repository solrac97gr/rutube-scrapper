package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

type Influencer struct {
	Name      string
	Followers string
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

	// Scrape the links concurrently
	influencers := scrapeLinksConcurrently(links)

	// Save the results to an Excel file
	err = saveToExcel("influencers.xlsx", influencers)
	if err != nil {
		log.Fatalf("Failed to save to Excel: %v", err)
	}

	fmt.Println("Scraping complete! Results saved to 'influencers.xlsx'")
}

func scrapeRutubeProfile(url string) (Influencer, error) {
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

	// Simplify the selectors based on observed structure
	name := doc.Find(".wdp-feed-banner-module__wdp-feed-banner__title h1 span").Text()
	name = strings.TrimSpace(name)

	followers := doc.Find(".wdp-feed-banner-module__wdp-feed-banner__title p").Text()
	followers = strings.TrimSpace(followers)

	// Clean up the followers string to remove "подписчиков" and keep only the number
	re := regexp.MustCompile(`[^\d]`)              // Matches any non-digit characters
	followers = re.ReplaceAllString(followers, "") // Removes non-digit characters

	if name == "" || followers == "" {
		return Influencer{}, fmt.Errorf("failed to extract name or followers, check selectors")
	}

	return Influencer{Name: name, Followers: followers}, nil
}

func scrapeLinksConcurrently(links []string) []Influencer {
	var wg sync.WaitGroup
	var mu sync.Mutex
	influencers := []Influencer{}

	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			influencer, err := scrapeRutubeProfile(link)
			if err != nil {
				log.Printf("Error scraping %s: %v", link, err)
				return
			}
			mu.Lock()
			influencers = append(influencers, influencer)
			mu.Unlock()
		}(link)
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
