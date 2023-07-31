package embed_test

import (
	"bytes"
	"testing"

	embed "goldmark-embed"
	"github.com/yuin/goldmark"
)

func TestMeta(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			embed.New(),
		),
	)
	source := `# Hello goldmark-embed

![](https://www.youtube.com/watch?v=dQw4w9WgXcQ)
`
	var buf bytes.Buffer
	if err := markdown.Convert([]byte(source), &buf); err != nil {
		panic(err)
	}
	if buf.String() != `<h1>Hello goldmark-embed</h1>
<p><iframe class="embeded-video youtube-embeded-video" width="100%" height="400" src="https://www.youtube.com/embed/dQw4w9WgXcQ" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe></p>
` {
		t.Error("Invalid HTML output")
		t.Log(buf.String())
	}
}
