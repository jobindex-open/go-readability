package readability

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	nurl "net/url"
	"strings"
	"time"

	"codeberg.org/readeck/go-readability/render"
	"github.com/araddon/dateparse"
	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

// Parse parses a reader and find the main readable content.
func (ps *Parser) Parse(input io.Reader, pageURL *nurl.URL) (Article, error) {
	// Parse input
	doc, err := dom.Parse(input)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse input: %v", err)
	}

	return ps.ParseAndMutate(doc, pageURL)
}

// ParseDocument parses the specified document and find the main readable content.
func (ps *Parser) ParseDocument(doc *html.Node, pageURL *nurl.URL) (Article, error) {
	// Clone document to make sure the original kept untouched
	return ps.ParseAndMutate(dom.Clone(doc, true), pageURL)
}

// ParseAndMutate is like ParseDocument, but mutates doc during parsing.
func (ps *Parser) ParseAndMutate(doc *html.Node, pageURL *nurl.URL) (Article, error) {
	// Backward compatibility with old logging approach.
	if ps.Debug {
		ps.Logger = newLegacyLogger(log.Default().Writer())
	}

	ps.doc = doc

	// Reset parser data
	ps.articleTitle = ""
	ps.articleByline = ""
	ps.articleDir = ""
	ps.articleSiteName = ""
	ps.documentURI = pageURL
	ps.attempts = []parseAttempt{}
	// These flags could get modified during subsequent passes in grabArticle
	ps.flags = flags{
		stripUnlikelys:     true,
		useWeightClasses:   true,
		cleanConditionally: true,
	}

	// Avoid parsing too large documents, as per configuration option
	if ps.MaxElemsToParse > 0 {
		numTags := len(dom.GetElementsByTagName(ps.doc, "*"))
		if numTags > ps.MaxElemsToParse {
			return Article{}, fmt.Errorf("documents too large: %d elements", numTags)
		}
	}

	// Unwrap image from noscript
	ps.unwrapNoscriptImages(ps.doc)

	// Extract JSON-LD metadata before removing scripts
	var jsonLd map[string]string
	if !ps.DisableJSONLD {
		jsonLd, _ = ps.getJSONLD()
	}

	// Remove script tags from the document.
	ps.removeScripts(ps.doc)

	// Prepares the HTML document
	ps.prepDocument()

	// Fetch metadata
	metadata := ps.getArticleMetadata(jsonLd)
	ps.articleTitle = metadata["title"]
	ps.articleByline = metadata["byline"]

	// Try to grab article content
	finalHTMLContent := ""
	finalTextContent := ""
	articleContent := ps.grabArticle()
	var readableNode *html.Node

	if articleContent != nil {
		ps.postProcessContent(articleContent)

		// If we haven't found an excerpt in the article's metadata,
		// use the article's first paragraph as the excerpt. This is used
		// for displaying a preview of the article's content.
		if metadata["excerpt"] == "" {
			if paragraph := getElementByTagName(articleContent, "p"); paragraph != nil {
				metadata["excerpt"] = strings.TrimSpace(render.InnerText(paragraph))
			}
		}

		readableNode = dom.FirstElementChild(articleContent)
		finalHTMLContent = dom.InnerHTML(articleContent)
		finalTextContent = render.InnerText(articleContent)
		finalTextContent = strings.TrimSpace(finalTextContent)
	}

	// Excerpt is an supposed to be short and concise,
	// so it shouldn't have any new line
	excerpt := strings.TrimSpace(metadata["excerpt"])
	excerpt = strings.Join(strings.Fields(excerpt), " ")

	// go-readability special:
	// Internet is dangerous and weird, and sometimes we will find
	// metadata isn't encoded using a valid Utf-8, so here we check it.
	var replacementTitle string
	if pageURL != nil {
		replacementTitle = pageURL.String()
	}

	validTitle := strings.ToValidUTF8(ps.articleTitle, replacementTitle)
	validByline := strings.ToValidUTF8(ps.articleByline, "")
	validExcerpt := strings.ToValidUTF8(excerpt, "")

	publishedTime := ps.getDate(metadata, "publishedTime")
	modifiedTime := ps.getDate(metadata, "modifiedTime")

	return Article{
		Title:         validTitle,
		Byline:        validByline,
		Node:          readableNode,
		Content:       finalHTMLContent,
		TextContent:   finalTextContent,
		Length:        charCount(finalTextContent),
		Excerpt:       validExcerpt,
		SiteName:      metadata["siteName"],
		Image:         metadata["image"],
		Favicon:       metadata["favicon"],
		Language:      ps.articleLang,
		PublishedTime: publishedTime,
		ModifiedTime:  modifiedTime,
	}, nil
}

// getDate tries to get a date from metadata, and parse it using a list of known formats.
func (ps *Parser) getDate(metadata map[string]string, fieldName string) *time.Time {
	dateStr, ok := metadata[fieldName]
	if !ok || len(dateStr) == 0 {
		return nil
	}
	d, err := dateparse.ParseAny(dateStr)
	if err != nil {
		ps.Logger.Warn("failed to parse timestamp",
			slog.Group("metadata",
				slog.String("field", fieldName),
				slog.String("value", dateStr),
			),
			slog.Any("err", err),
		)
		return nil
	}
	return &d
}

func newLegacyLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(
		w,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "time" {
					return slog.Attr{}
				}
				if a.Value.Kind() == slog.KindFloat64 {
					return slog.String(a.Key, fmt.Sprintf("%.2f", a.Value.Float64()))
				}
				return a
			},
		},
	))
}
