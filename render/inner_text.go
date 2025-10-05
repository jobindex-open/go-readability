package render

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

const noBreakSpace = '\u00A0'

// Render plain text from HTML DOM with awareness of block-level elements
// https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/innerText
func InnerText(doc *html.Node) string {
	var tb innerTextBuilder

	var render func(*html.Node, bool)
	render = func(n *html.Node, keepWhitespace bool) {
		if n.Type == html.TextNode {
			if keepWhitespace {
				tb.WritePre(n.Data)
			} else {
				// write each word to innerTextBuilder
				startOfWord := -1
				for i, r := range n.Data {
					if unicode.IsSpace(r) {
						if startOfWord >= 0 {
							tb.WriteWord(n.Data[startOfWord:i])
							startOfWord = -1
						}
						if r == noBreakSpace {
							tb.QueueSpace(noBreakSpace)
						} else {
							tb.QueueSpace(' ')
						}
					} else if startOfWord < 0 {
						startOfWord = i
					}
				}
				if startOfWord >= 0 {
					tb.WriteWord(n.Data[startOfWord:])
				}
			}
			return
		}

		if n.Type == html.ElementNode {
			switch n.Data {
			// These elements will never contain user-facing text nodes, so there is no need to
			// recurse into them.
			case "head", "meta", "style", "script", "iframe",
				"audio", "video", "track", "source", "canvas", "svg", "map", "area":
				return
			case "br":
				tb.WriteNewline(1, false)
			case "hr", "p", "blockquote", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "dl", "table":
				tb.WriteNewline(2, true)
			case "pre":
				tb.WriteNewline(2, true)
				keepWhitespace = true
			case "th", "td":
				tb.QueueSpace('\t')
			case "div", "figure", "figcaption", "picture", "li", "dt", "dd",
				"header", "footer", "main", "section", "article", "aside", "nav", "address",
				"details", "summary", "dialog", "form", "fieldset",
				"caption", "thead", "tbody", "tfoot", "tr":
				tb.WriteNewline(1, true)
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			render(child, keepWhitespace)
		}
	}

	render(doc, false)
	return tb.String()
}

type innerTextBuilder struct {
	sb strings.Builder
	sp rune
	nl uint8
}

func (tb *innerTextBuilder) String() string {
	return tb.sb.String()
}

func (tb *innerTextBuilder) QueueSpace(c rune) {
	if tb.sp > 0 {
		return
	}
	tb.sp = c
}

func (tb *innerTextBuilder) WriteNewline(n uint8, collapse bool) {
	if collapse {
		if tb.nl >= n {
			return
		}
		n -= tb.nl
	}
	tb.nl += n
	if collapse && tb.sb.Len() == 0 {
		return
	}
	for ; n > 0; n-- {
		tb.sb.WriteByte('\n')
	}
}

func (tb *innerTextBuilder) WriteWord(w string) {
	if tb.sp > 0 && tb.nl == 0 {
		tb.sb.WriteRune(tb.sp)
	}
	tb.sb.WriteString(w)
	tb.nl = 0
	tb.sp = 0
}

func (tb *innerTextBuilder) WritePre(pre string) {
	tb.sb.WriteString(pre)
	tb.nl = 0
	tb.sp = 0
}
