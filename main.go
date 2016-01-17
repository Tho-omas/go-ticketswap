package main

import (
	"net/http"

	"io"

	"net/url"

	"strconv"

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
	Title string
	Price float64
	User  string
	Url   *url.URL
}

const (
	tagSection = "section"
	tagDiv     = "div"
	tagH2      = "h2"
	tagA       = "a"
	tagArticle = "article"
	attrProp   = "itemprop"
	attrClass  = "class"
	attrHref   = "href"

	adsAttrClassVal = "ad-list"
	availableData   = "Available"

	ticketAttrPropVal       = "tickets"
	ticketUrlAttrPropVal    = "offerurl"
	ticketTitleAttrClassVal = "ad-list-title"
	ticketPriceAttrClassVal = "ad-list-price"
)

func parseHtml(r io.Reader) []Ticket {
	tokenizer := html.NewTokenizer(r)
	var tickets []Ticket
	var t *Ticket
	for {
		token := tokenizer.Token()
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			// End of the document, we're done
			return tickets
		case html.StartTagToken:
			if token.Data == tagSection { // ads section found
				tokenizer.Next()
				// found some ads, check for availability

			} else if token.Data == tagArticle && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketAttrPropVal}) {
				t = &Ticket{}
			} else if token.Data == tagA && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketUrlAttrPropVal}) {
				urlString, _ := getVal(token.Attr, attrHref)
				url, _ := url.Parse(urlString)
				t.Url = url
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketTitleAttrClassVal}) {
				tokenizer.Next()
				t.Title = tokenizer.Token().Data
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketPriceAttrClassVal}) {
				tokenizer.Next()
				price, _ := strconv.ParseFloat(tokenizer.Token().Data, 64)
				t.Price = price
			}
			// if ok, ad := getAd(tokenizer); ok {
			// 	fmt.Println(ad.Desc, url)
			// 	return
			// }
		case html.TextToken:
		case html.EndTagToken:
			if tokenizer.Token().Data == tagArticle && t != nil {
				tickets = append(tickets, *t)
				t = nil
			}

		}
	}
	return tickets
}

func hasAttr(attrs []html.Attribute, attr *html.Attribute) bool {
	for _, a := range attrs {
		if a.Key == attr.Key && a.Val == attr.Val {
			return true
		}
	}
	return false
}

func getVal(attrs []html.Attribute, key string) (string, error) {
	for _, a := range attrs {
		if a.Key == key {
			return a.Val, nil
		}
	}
	return "", nil
}
