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

// Advertisement defines the properties of an advertisement from the ticketswap.com.
type Advertisement struct {
	Title    string
	Qty      uint8
	Price    float64
	Currency string
	User     string
	Url      url.URL
}

// String is part of fmt.Stringer.
func (a Advertisement) String() (res string) {
	return fmt.Sprintf("%d x %.2f %s by %s %s", a.Qty, a.Price, a.Currency, a.User, a.Url.String())
}

// Advertisements defines a slice of advertisements.
type Advertisements []Advertisement

// Len is part of sort.Interface.
func (a Advertisements) Len() int { return len(a) }

// Less is part of sort.Interface. It is implemented by comparing the ads prices.
func (a Advertisements) Less(i, j int) bool { return a[i].Price < a[j].Price }

// Swap is part of sort.Interface.
func (a Advertisements) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// String is part of fmt.Stringer.
// Returns the string representation of the Advertisements instance.
func (a Advertisements) String() string {
	var buffer bytes.Buffer
	for _, ad := range a {
		buffer.WriteString(ad.String())
		buffer.WriteString("\n")
	}
	return buffer.String()
}

// NewAdvertisements does try to build an ads slice from the html.
// Returns an error if it fails to parse ads from the input reader.
func NewAdvertisements(r io.Reader) (Advertisements, error) {
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
			return ads, nil
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
				urlPath, err := getVal(token.Attr, attrHref)
				if err != nil {
					return nil, err
				}
				url, err := url.Parse(path.Join(path.Base(baseUrl), urlPath))
				if err != nil {
					return nil, err
				}
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
				priceStr, err := getVal(token.Attr, attrContent)
				if err != nil {
					return nil, err
				}
				price, err := strconv.ParseFloat(priceStr, 64)
				if err != nil {
					return nil, err
				}
				ad.Price = price
			} else if token.Data == tagMeta && hasAttr(token.Attr, &html.Attribute{Key: attrProp, Val: adUrlQuantityPropVal}) {
				qtyStr, err := getVal(token.Attr, attrContent)
				if err != nil {
					return nil, err
				}
				qty, err := strconv.ParseUint(qtyStr, 10, 8)
				if err != nil {
					return nil, err
				}
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
	return ads, nil
}
