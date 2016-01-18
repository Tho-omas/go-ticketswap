package ticketswap

import (
	"io"
	"net/url"
	"path"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

const (
	ticketsAttrClassVal = "type-list"

	ticketTitleAttrClassVal = "type-list-title"
	ticketDateAttrClassVal  = "type-list-date"
	ticketCountAttrClassVal = "tickets-count"
)

type Ticket struct {
	Title string
	Date  time.Time
	Qty   uint16
	Url   url.URL
}

type Tickets []Ticket

func (a Tickets) Len() int           { return len(a) }
func (a Tickets) Less(i, j int) bool { return a[i].Title < a[j].Title }
func (a Tickets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func NewTickets(r io.Reader) Tickets {
	tokenizer := html.NewTokenizer(r)
	var tickets Tickets
	var t *Ticket
	didFoundTickets := false
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			// End of the document, we're done
			return tickets
		case html.StartTagToken:
			if token.Data == tagSection && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketsAttrClassVal}) {
				didFoundTickets = true
			}
			if !didFoundTickets {
				continue
			} else if token.Data == tagArticle {
				tokenizer.Next()
				tokenizer.Next()
				token := tokenizer.Token()
				if token.Data == tagA {
					t = &Ticket{}
					urlPath, _ := getVal(token.Attr, attrHref)
					url, _ := url.Parse(path.Join(path.Base(baseUrl), urlPath))
					t.Url = *url
				}
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketTitleAttrClassVal}) {
				tokenizer.Next()
				t.Title = tokenizer.Token().Data
			} else if token.Data == tagDiv && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketDateAttrClassVal}) {
				tokenizer.Next()
				layout := "Monday, January 2, 2006"
				date, _ := time.Parse(layout, tokenizer.Token().Data)
				t.Date = date
			} else if token.Data == tagSpan && hasAttr(token.Attr, &html.Attribute{Key: attrClass, Val: ticketCountAttrClassVal}) {
				tokenizer.Next()
				qty, _ := strconv.ParseUint(tokenizer.Token().Data, 10, 16)
				t.Qty = uint16(qty)
			}
		case html.EndTagToken:
			if token.Data == tagArticle && didFoundTickets && t != nil {
				tickets = append(tickets, *t)
				t = nil
			} else if token.Data == tagSection {
				didFoundTickets = false
			}
		}
	}
	return tickets
}
