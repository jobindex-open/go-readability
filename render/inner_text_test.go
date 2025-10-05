package render

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestInnerText(t *testing.T) {
	tests := []struct {
		name string
		html string
		text string
	}{
		{
			name: "mixed",
			html: `<p><span>Hi there!<p><img/>How have you been?
<ul><li>pretty good</li><li>not bad</li><li>meh</li>
<p><span>Inline</span><span>No</span><span>Spaces</span><div>Fin
<table><tr><th>header 1</th><th>header 2</th></tr><tr><td>cell 1</td><td>cell 2</td></tr></table>
<div>
	<div>
		<div>
			<div>
				Deeply nested
			</div>
		</div>
	</div>
</div>`,
			text: `Hi there!

How have you been?

pretty good
not bad
meh

InlineNoSpaces
Fin

header 1	header 2
cell 1	cell 2
Deeply nested`,
		},
		{
			name: "no break space",
			html: `<div> <p> open&nbsp;source software </p> </div>`,
			text: "open\u00a0source software",
		},
		{
			name: "br element",
			html: `<p>hard<br>line<br><br><br>breaks</p>`,
			text: "hard\nline\n\n\nbreaks",
		},
		{
			name: "pre element",
			html: `<div>
	<p> Example code: </p>
	<pre><code>def normalize(s: str) -&gt; str:
    <span class="comment"># remove all U+00AD (SOFT HYPHEN)</span>
    return s.<span class="fn">replace</span>('\u00ad', '')
</code></pre>
</div>`,
			text: `Example code:

def normalize(s: str) -> str:
    # remove all U+00AD (SOFT HYPHEN)
    return s.replace('\u00ad', '')
`,
		},
		{
			name: "headings",
			html: `<h1>HEADING 1</h1>
	<p>First paragraph</p>
	<h2>HEADING 2</h2>
	<p>Second paragraph</p>
`,
			text: `HEADING 1

First paragraph

HEADING 2

Second paragraph`,
		},
		{
			name: "multibyte",
			html: `<p align="center">
	<a href="../../../index.html">福娘童話集</a> &gt; <a href="../index.html">きょうのイソップ童話</a> &gt; <a href="../itiran/01gatu.htm">１月のイソップ童話</a> &gt; 欲張りなイヌ
</p>`,
			text: `福娘童話集 > きょうのイソップ童話 > １月のイソップ童話 > 欲張りなイヌ`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}

			gotText := InnerText(doc)
			if gotText != tt.text {
				t.Errorf("mismatched text:\nwant: %q\ngot:  %q\n", tt.text, gotText)
			}
		})
	}
}
