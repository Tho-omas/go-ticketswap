package main

import (
	"net/http"
	"strconv"

	"io"

	"net/url"

	"fmt"

	"path"

	"sort"

	"golang.org/x/net/html"
)

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// 1. fetch html
	url := "https://www.ticketswap.com/event/rihanna/floor/c1671553-db2b-4f0f-b9c1-51a70e6b48e0/4857"
	response, err := http.Get(url)
	perror(err)

	// bytes, _ := ioutil.ReadAll(response.Body)
	// xml := string(bytes)

	// 2. parse html
	defer response.Body.Close()
	tickets := parseHtml(response.Body)
	sort.Sort(tickets)
	fmt.Println(tickets)
}

type Ticket struct {
	Title    string
	Qty      uint8
	Price    float64
	Currency string
	User     string
	Url      *url.URL
}

func (t *Ticket) String() string {
	return fmt.Sprintf("%s, %d x %.2f %s by %s, %s", t.Title, t.Qty, t.Price, t.Currency, t.User, t.Url.String())
}

type Tickets []Ticket

func (a Tickets) Len() int           { return len(a) }
func (a Tickets) Less(i, j int) bool { return a[i].Price < a[j].Price }
func (a Tickets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

const (
	baseUrl     = "https://www.ticketswap.com/"
	tagSection  = "section"
	tagDiv      = "div"
	tagH2       = "h2"
	tagA        = "a"
	tagMeta     = "meta"
	tagArticle  = "article"
	attrProp    = "itemprop"
	attrClass   = "class"
	attrHref    = "href"
	attrContent = "content"

	adsAttrClassVal = "ad-list"
	availableData   = "Available"

	ticketAttrPropVal        = "tickets"
	ticketUrlAttrPropVal     = "offerurl"
	ticketUrlQuantityPropVal = "quantity"
	ticketUrlPricePropVal    = "price"
	ticketUrlCurrencyPropVal = "currency"
	ticketUserAttrClassVal   = "name"
	ticketTitleAttrClassVal  = "ad-list-title"
	ticketPriceAttrClassVal  = "ad-list-price"
)

func parseHtml(r io.Reader) Tickets {
	tokenizer := html.NewTokenizer(r)
	var tickets Tickets
	var t *Ticket
	isAvailable := false
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			// End of the document, we're done
			return tickets
		case html.StartTagToken:
			if token.Data == tagSection && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: adsAttrClassVal}) {
				// found some ads, check for availability
				tokenizer.Next()
				tokenizer.Next()
				tokenizer.Next()
				isAvailable = tokenizer.Token().Data == availableData
			}
			if !isAvailable {
				continue
			} else if token.Data == tagArticle && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketAttrPropVal}) {
				t = &Ticket{}
			} else if token.Data == tagA && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketUrlAttrPropVal}) {
				urlPath, _ := getVal(token.Attr, attrHref)
				url, _ := url.Parse(path.Join(path.Base(baseUrl), urlPath))
				t.Url = url
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketTitleAttrClassVal}) {
				tokenizer.Next()
				t.Title = tokenizer.Token().Data
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketUserAttrClassVal}) {
				tokenizer.Next()
				t.User = tokenizer.Token().Data
			}
		case html.TextToken:
		case html.SelfClosingTagToken:
			if !isAvailable {
				continue
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketUrlPricePropVal}) {
				priceStr, _ := getVal(token.Attr, attrContent)
				price, _ := strconv.ParseFloat(priceStr, 64)
				t.Price = price
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketUrlQuantityPropVal}) {
				qtyStr, _ := getVal(token.Attr, attrContent)
				qty, _ := strconv.ParseUint(qtyStr, 10, 8)
				t.Qty = uint8(qty)
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: ticketUrlCurrencyPropVal}) {
				currencyStr, _ := getVal(token.Attr, attrContent)
				t.Currency = currencyStr
			}
		case html.EndTagToken:
			if token.Data == tagArticle && isAvailable && t != nil {
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
