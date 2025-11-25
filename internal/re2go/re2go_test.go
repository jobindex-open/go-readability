package re2go

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsByline(t *testing.T) {
	tests := map[string]bool{
		"article-byline":   true,
		"author":           true,
		"dateline":         true,
		"meta-prep-author": true,
		"writtenbynames":   true,

		"article-line": false,
		"autor":        false,
		"bynames":      false,
		"date":         false,
		"meta-autor":   false,
	}
	for matchString, expected := range tests {
		if got := IsByline(matchString); got != expected {
			t.Errorf("IsByline(%q): expected %v, got %v", matchString, expected, got)
		}
		upper := strings.ToUpper(matchString)
		if got := IsByline(upper); got != expected {
			t.Errorf("IsByline(%q): expected %v, got %v", upper, expected, got)
		}
	}
}

func Test_IsPositiveClass(t *testing.T) {
	tests := map[string]bool{
		"article":    true,
		"blog":       true,
		"body":       true,
		"content":    true,
		"entry":      true,
		"h-entry":    true,
		"hentry":     true,
		"main":       true,
		"page":       true,
		"pagination": true,
		"post":       true,
		"story":      true,
		"text":       true,

		"container":   false,
		"description": false,
		"footer":      false,
		"gallery":     false,
		"header":      false,
		"layout":      false,
		"navigation":  false,
		"news":        false,
		"sidebar":     false,
		"toolbar":     false,
		"widget":      false,
	}
	for matchString, expected := range tests {
		if got := IsPositiveClass(matchString); got != expected {
			t.Errorf("IsPositiveClass(%q): expected %v, got %v", matchString, expected, got)
		}
		upper := strings.ToUpper(matchString)
		if got := IsPositiveClass(upper); got != expected {
			t.Errorf("IsPositiveClass(%q): expected %v, got %v", upper, expected, got)
		}
	}
}

func Test_IsNegativeClass(t *testing.T) {
	tests := map[string]bool{
		"-ad-":           true,
		"ad-banner":      true,
		"banner":         true,
		"class hid good": true,
		"class hid":      true,
		"com-":           true,
		"combx":          true,
		"comment":        true,
		"contact":        true,
		"footer":         true,
		"gdpr":           true,
		"hid class":      true,
		"hid":            true,
		"hidden":         true,
		"masthead":       true,
		"media":          true,
		"meta":           true,
		"outbrain":       true,
		"promo":          true,
		"related":        true,
		"scroll":         true,
		"share":          true,
		"shopping":       true,
		"shoutbox":       true,
		"sidebar":        true,
		"skyscraper":     true,
		"sponsor":        true,
		"tags":           true,
		"widget":         true,

		"catalog":   false,
		"details":   false,
		"foot":      false,
		"footnote":  false,
		"gallery":   false,
		"navbar":    false,
		"news-feed": false,
		"overview":  false,
		"profile":   false,
		"support":   false,
		"tool":      false,
		"toolbar":   false,
		"user-menu": false,
	}
	for matchString, expected := range tests {
		if got := IsNegativeClass(matchString); got != expected {
			t.Errorf("IsNegativeClass(%q): expected %v, got %v", matchString, expected, got)
		}
		upper := strings.ToUpper(matchString)
		if got := IsNegativeClass(upper); got != expected {
			t.Errorf("IsNegativeClass(%q): expected %v, got %v", upper, expected, got)
		}
	}
}

func Test_IsUnlikelyCandidates(t *testing.T) {
	tests := map[string]bool{
		"-ad-":         true,
		"ad-banner":    true,
		"ad-break":     true,
		"agegate":      true,
		"ai2html":      true,
		"banner":       true,
		"breadcrumbs":  true,
		"combx":        true,
		"comment":      true,
		"community":    true,
		"cover-wrap":   true,
		"disqus":       true,
		"extra":        true,
		"footer":       true,
		"gdpr":         true,
		"header":       true,
		"legends":      true,
		"menu":         true,
		"pager":        true,
		"pagination":   true,
		"popup":        true,
		"related":      true,
		"remark":       true,
		"replies":      true,
		"rss":          true,
		"shoutbox":     true,
		"sidebar":      true,
		"skyscraper":   true,
		"social":       true,
		"sponsor":      true,
		"supplemental": true,
		"yom-remote":   true,

		"catalog":   false,
		"container": false,
		"gallery":   false,
		"newsfeed":  false,
		"overview":  false,
		"summary":   false,
		"toolbar":   false,
	}
	for matchString, expected := range tests {
		if got := IsUnlikelyCandidates(matchString); got != expected {
			t.Errorf("IsUnlikelyCandidates(%q): expected %v, got %v", matchString, expected, got)
		}
		upper := strings.ToUpper(matchString)
		if got := IsUnlikelyCandidates(upper); got != expected {
			t.Errorf("IsUnlikelyCandidates(%q): expected %v, got %v", upper, expected, got)
		}
	}
}

func Test_MaybeItsACandidate(t *testing.T) {
	tests := map[string]bool{
		"and":     true,
		"article": true,
		"body":    true,
		"column":  true,
		"content": true,
		"main":    true,
		"shadow":  true,

		"footer":  false,
		"gallery": false,
		"header":  false,
		"menu":    false,
		"navbar":  false,
		"text":    false,
	}
	for matchString, expected := range tests {
		if got := MaybeItsACandidate(matchString); got != expected {
			t.Errorf("MaybeItsACandidate(%q): expected %v, got %v", matchString, expected, got)
		}
		upper := strings.ToUpper(matchString)
		if got := MaybeItsACandidate(upper); got != expected {
			t.Errorf("MaybeItsACandidate(%q): expected %v, got %v", upper, expected, got)
		}
	}
}

func Test_NormalizeSpaces(t *testing.T) {
	assert.Equal(t, "some sentence", NormalizeSpaces("some   sentence"))
	assert.Equal(t, "with tabs", NormalizeSpaces("with \t \ttabs"))
	assert.Equal(t, " single space is ok ", NormalizeSpaces(" single space is ok "))
	assert.Equal(t, " multi space removed ", NormalizeSpaces("   multi   space   removed   "))
}
