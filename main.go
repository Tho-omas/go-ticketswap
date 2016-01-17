package main

import (
	"fmt"
	"net/http"

	"io"

	"net/url"

	"golang.org/x/net/html"
)

type Ad struct {
	Desc  string
	Price string
}

func perror(err error) {
	if err != nil {
		print("Error!!!")
		panic(err)
	}
}

func getAd(tokenizer *html.Tokenizer) (ok bool, ad Ad) {
	t := tokenizer.Token()

	isDiv := t.Data == "div"
	if !isDiv {
		return
	}

	for _, a := range t.Attr {
		if a.Key == "class" && a.Val == "ad-list-title" {
			tokenizer.Next()
			title := tokenizer.Token().Data
			tokenizer.Next()
			tokenizer.Next()
			tokenizer.Next()
			tokenizer.Next()
			isSold := tokenizer.Token().Data == "Sold"
			if !isSold {
				ok = true
				ad = Ad{Desc: title}
				return
			}
		}
	}
	return
}

func main() {
	// 1. fetch html
	url := "https://www.ticketswap.com/event/rihanna/floor/c1671553-db2b-4f0f-b9c1-51a70e6b48e0/4857"
	response, err := http.Get(url)
	perror(err)

	// bytes, _ := ioutil.ReadAll(response.Body)
	// xml := string(bytes)

	// 2. parse html
	parseHtml(response.Body)
	defer response.Body.Close()
}

type Ticket struct {
	Description string
	Price       string
	User        string
	Url         url.URL
}

func parseHtml(r io.Reader) {
	tokenizer := html.NewTokenizer(r)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			// End of the document, we're done
			return
		case html.StartTagToken:
			if ok, ad := getAd(tokenizer); ok {
				fmt.Println(ad.Desc, url)
				return
			}
		case html.TextToken:
		case html.EndTagToken:
		}

	}
}
