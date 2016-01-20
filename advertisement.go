package ticketswap

import (
	"bytes"
	"io"
	"net/url"
	"path"
	"strconv"

	"fmt"

	"golang.org/x/net/html"
)

const (
	adsAttrClassVal = "ad-list"
	availableData   = "Available"

	adTicketAttrPropVal  = "tickets"
	adUrlAttrPropVal     = "offerurl"
	adUrlQuantityPropVal = "quantity"
	adUrlPricePropVal    = "price"
	adUrlCurrencyPropVal = "currency"
	adUserAttrClassVal   = "name"
	adTitleAttrClassVal  = "ad-list-title"
	adPriceAttrClassVal  = "ad-list-price"
)

type Advertisement struct {
	Title    string
	Qty      uint8
	Price    float64
	Currency string
	User     string
	Url      url.URL
}

func (a Advertisement) String() (res string) {
	return fmt.Sprintf("%d x %.2f %s by %s %s", a.Qty, a.Price, a.Currency, a.User, a.Url.String())
}

type Advertisements []Advertisement

func (a Advertisements) Len() int           { return len(a) }
func (a Advertisements) Less(i, j int) bool { return a[i].Price < a[j].Price }
func (a Advertisements) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Advertisements) String() string {
	var buffer bytes.Buffer
	for _, ad := range a {
		buffer.WriteString(ad.String())
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func NewAdvertisements(r io.Reader) Advertisements {
	tokenizer := html.NewTokenizer(r)
	var ads Advertisements
	var ad *Advertisement
	isAvailable := false
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			// End of the document, we're done
			return ads
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
			} else if token.Data == tagArticle && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: adTicketAttrPropVal}) {
				ad = &Advertisement{}
			} else if token.Data == tagA && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: adUrlAttrPropVal}) {
				urlPath, _ := getVal(token.Attr, attrHref)
				url, _ := url.Parse(path.Join(path.Base(baseUrl), urlPath))
				ad.Url = *url
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: adTitleAttrClassVal}) {
				tokenizer.Next()
				ad.Title = tokenizer.Token().Data
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: adUserAttrClassVal}) {
				tokenizer.Next()
				ad.User = tokenizer.Token().Data
			}
		case html.SelfClosingTagToken:
			if !isAvailable {
				continue
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: adUrlPricePropVal}) {
				priceStr, _ := getVal(token.Attr, attrContent)
				price, _ := strconv.ParseFloat(priceStr, 64)
				ad.Price = price
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: adUrlQuantityPropVal}) {
				qtyStr, _ := getVal(token.Attr, attrContent)
				qty, _ := strconv.ParseUint(qtyStr, 10, 8)
				ad.Qty = uint8(qty)
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: adUrlCurrencyPropVal}) {
				currencyStr, _ := getVal(token.Attr, attrContent)
				ad.Currency = currencyStr
			}
		case html.EndTagToken:
			if token.Data == tagArticle && isAvailable && ad != nil {
				ads = append(ads, *ad)
				ad = nil
			}
		}
	}
	return ads
}
