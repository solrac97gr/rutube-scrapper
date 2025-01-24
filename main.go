package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

	// Process each link and extract data
	var influencers []Influencer
	for _, link := range links {
		fmt.Printf("Scraping: %s\n", link)
		influencer, err := scrapeRutubeProfile(link)
		if err != nil {
			log.Printf("Failed to scrape %s: %v", link, err)
			continue
		}
		influencers = append(influencers, influencer)
	}

	// Print the results
	fmt.Println("Results:")
	for _, influencer := range influencers {
		fmt.Printf("Name: %s, Followers: %s\n", influencer.Name, influencer.Followers)
	}
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

	if name == "" || followers == "" {
		return Influencer{}, fmt.Errorf("failed to extract name or followers, check selectors")
	}

	return Influencer{Name: name, Followers: followers}, nil
}
