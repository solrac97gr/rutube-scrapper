# Rutube Scraper

Rutube Scraper is a Go application with a graphical user interface (GUI) that allows users to scrape influencer data from Rutube profiles. The application extracts the name and number of followers of influencers from a list of URLs provided in a `influencers.txt` file. The results are saved in an Excel file (`influencers.xlsx`) at a location specified by the user.

## Features

- **File Upload:** Easily select the `influencers.txt` file containing a list of Rutube profile URLs.
- **Web Scraping:** Extracts influencer names and follower counts from Rutube profiles.
- **Excel Export:** Saves the scraped data into an Excel file with the order of URLs preserved.
- **User-Friendly GUI:** Simple interface for uploading and saving files.

## Requirements

- Go 1.18 or higher
- Internet connection (for scraping Rutube profiles)

### Dependencies

The application uses the following Go packages:

- [fyne](https://fyne.io/) for the GUI.
- [goquery](https://github.com/PuerkitoBio/goquery) for web scraping.
- [excelize](https://github.com/xuri/excelize) for creating Excel files.

Install dependencies using:
```bash
go get fyne.io/fyne/v2
go get github.com/PuerkitoBio/goquery
go get github.com/xuri/excelize/v2
```

## How to Use

1. Clone this repository:
   ```bash
   git clone https://github.com/your-username/rutube-scraper.git
   cd rutube-scraper
   ```

2. Create a `influencers.txt` file with Rutube profile URLs, each on a new line:
   ```txt
   https://rutube.ru/channel/12345/
   https://rutube.ru/channel/67890/
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

4. Use the GUI to:
   - Upload the `influencers.txt` file.
   - Save the resulting `influencers.xlsx` file to your desired location.

## Output Format

The generated Excel file (`influencers.xlsx`) contains two columns:

| Name          | Followers |
|---------------|-----------|
| Viki Show     | 17214     |
| Super Papa    | 8776      |

## Notes

- Ensure the `influencers.txt` file contains valid Rutube profile URLs.
- The application processes profiles in the order they appear in the file.
- If a profile's data cannot be scraped, it will be skipped, and an error will be logged in the console.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests to improve the application.
