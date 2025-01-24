all: build

build:
	@go build -o rutube-scrapper main.go

pack:
	@go install fyne.io/fyne/v2/cmd/fyne@latest
	@fyne package -os darwin -icon icon.jpeg

clean:
	@rm -f rutube-scrapper
