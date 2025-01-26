package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

type Influencer struct {
	Name      string
	Followers string
	Index     int
}

func main() {
	// Create a Fyne application
	a := app.New()
	w := a.NewWindow("Rutube Scraper for Nastya ❤️")

	// Widgets
	title := widget.NewLabel("Rutube Scraper for Nastya ❤️")
	label := widget.NewLabel("Upload the influencers.txt file to start scraping.")
	resultLabel := widget.NewLabel("")
	startButton := widget.NewButton("Upload influencers.txt", func() {
		dialog.ShowFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if file == nil {
				return
			}
			links, err := readLinksFromFile(file)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Scrape links and allow saving the result
			influencers := scrapeLinksConcurrently(links)
			sort.Slice(influencers, func(i, j int) bool {
				return influencers[i].Index < influencers[j].Index
			})

			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				if writer == nil {
					return
				}
				defer writer.Close()

				err = saveToExcel(writer.URI().Path(), influencers)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				resultLabel.SetText("Scraping complete! Results saved successfully.")
			}, w)
		}, w)
	})

	// Layout
	content := container.NewVBox(
		title,
		label,
		startButton,
		resultLabel,
	)
	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 400))
	w.ShowAndRun()
}

func readLinksFromFile(file fyne.URIReadCloser) ([]string, error) {
	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		links = append(links, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	return links, nil
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

	// Extract name from the title attribute
	name := doc.Find("h1.wdp-feed-banner-module__wdp-feed-banner__title-text").AttrOr("title", "")
	name = strings.TrimSpace(name)

	// Extract followers
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
	influencers := make([]Influencer, len(links))
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
			influencers[i] = influencer
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
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), inf.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), inf.Followers)
	}

	// Save the Excel file
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
