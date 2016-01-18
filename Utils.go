package ticketswap

import "golang.org/x/net/html"

const (
	tagSection  = "section"
	tagDiv      = "div"
	tagH2       = "h2"
	tagA        = "a"
	tagMeta     = "meta"
	tagArticle  = "article"
	tagSpan     = "span"
	attrProp    = "itemprop"
	attrClass   = "class"
	attrHref    = "href"
	attrContent = "content"
)

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
