package readability

import (
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/net/html"
)

// inspectNode wraps a HTML node to use in structured logging
func inspectNode(node *html.Node) slog.LogValuer {
	return &inspectedNode{node}
}

type inspectedNode struct {
	node *html.Node
}

func (n *inspectedNode) LogValue() slog.Value {
	if n.node.Type == html.TextNode {
		return slog.StringValue(n.node.Data)
	}

	var tagPreview strings.Builder
	tagPreview.WriteString("<")
	tagPreview.WriteString(n.node.Data)

	hasOtherAttributes := false
	for _, attr := range n.node.Attr {
		switch strings.ToLower(attr.Key) {
		case "id", "class", "rel", "itemprop", "name", "type", "role", "for", "action", "method":
			fmt.Fprintf(&tagPreview, ` %s=%q`, attr.Key, attr.Val)
		case "src", "href":
			val := attr.Val
			if strings.HasPrefix(val, "data:") {
				if v, _, ok := strings.Cut(val, ","); ok {
					val = v + ",***"
				}
			} else if strings.HasPrefix(val, "javascript:") {
				val = "javascript:***"
			}
			fmt.Fprintf(&tagPreview, ` %s=%q`, attr.Key, val)
		default:
			if !strings.HasPrefix(attr.Key, "data-readability-") {
				hasOtherAttributes = true
			}
		}
	}
	if hasOtherAttributes {
		tagPreview.WriteString(" ...")
	}

	if n.node.FirstChild == nil {
		tagPreview.WriteString("/")
	}
	tagPreview.WriteString(">")
	return slog.StringValue(tagPreview.String())
}
