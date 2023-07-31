package embed

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Option is a functional option type for this extension.
type Option func(*embedExtension)

type embedExtension struct{}

const (
	EmbededVideoProviderYouTube  = "youtube"
	EmbededVideoProviderBilibili = "bilibili"
)

// New returns a new Embed extension.
func New(opts ...Option) goldmark.Extender {
	e := &embedExtension{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *embedExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(defaultASTTransformer, 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(), 500),
		),
	)
}

// EmbededVideo struct represents a EmbededVideo embed of the Markdown text.
type EmbededVideo struct {
	ast.Image
	Provider string
	VID      string
}

// KindEmbededVideo is a NodeKind of the YouTube node.
var KindEmbededVideo = ast.NewNodeKind("EmbededVideo")

// Kind implements Node.Kind.
func (n *EmbededVideo) Kind() ast.NodeKind {
	return KindEmbededVideo
}

// NewEmbededVideo returns a new YouTube node.
func NewEmbededVideo(img *ast.Image, provider, vid string) *EmbededVideo {
	c := &EmbededVideo{
		Image:    *img,
		Provider: provider,
		VID:      vid,
	}
	c.Destination = img.Destination
	c.Title = img.Title

	return c
}

type astTransformer struct{}

var defaultASTTransformer = &astTransformer{}

func (a *astTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	replaceImages := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Kind() != ast.KindImage {
			return ast.WalkContinue, nil
		}

		img := n.(*ast.Image)
		u, err := url.Parse(string(img.Destination))
		if err != nil {
			msg := ast.NewString([]byte(fmt.Sprintf("<!-- %s -->", err)))
			msg.SetCode(true)
			n.Parent().InsertAfter(n.Parent(), n, msg)
			return ast.WalkContinue, nil
		}

		// Embed a video?
		vid := ""
		provider := EmbededVideoProviderYouTube
		if u.Host == "www.youtube.com" && u.Path == "/watch" {
			// this is a youtube video: https://www.youtube.com/watch?v={vid}
			vid = u.Query().Get("v")
		} else if u.Host == "youtu.be" {
			// this is a youtube video too: https://youtu.be/{vid}
			vid = u.Path[1:]
			vid = strings.Trim(vid, "/")
		} else if u.Host == "www.bilibili.com" && strings.HasPrefix(u.Path, "/video/") {
			// this is a bilibili video: https://www.bilibili.com/video/{vid}
			vid = u.Path[7:]
			vid = strings.Trim(vid, "/")
			provider = EmbededVideoProviderBilibili
		} else {
			return ast.WalkContinue, nil
		}

		if vid != "" {
			ev := NewEmbededVideo(img, provider, vid)
			n.Parent().SetAttributeString("class", []byte("embeded-video-wrapper"))
			n.Parent().ReplaceChild(n.Parent(), n, ev)
		}

		return ast.WalkContinue, nil
	}

	ast.Walk(node, replaceImages)
}

// HTMLRenderer struct is a renderer.NodeRenderer implementation for the extension.
type HTMLRenderer struct{}

// NewHTMLRenderer builds a new HTMLRenderer with given options and returns it.
func NewHTMLRenderer() renderer.NodeRenderer {
	r := &HTMLRenderer{}
	return r
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindEmbededVideo, r.renderEmbededVideo)
}

func (r *HTMLRenderer) renderEmbededVideo(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}

	ev := node.(*EmbededVideo)
	if ev.Provider == EmbededVideoProviderYouTube {
		w.Write([]byte(`<iframe class="embeded-video youtube-embeded-video" width="100%" height="400" src="https://www.youtube.com/embed/` + ev.VID + `" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`))
	} else if ev.Provider == EmbededVideoProviderBilibili {
		w.Write([]byte(`<iframe class="embeded-video bilibili-embeded-video" width="100%" height="400" src="//player.bilibili.com/player.html?bvid=` + ev.VID + `&page=1" scrolling="no" border="0" framespacing="0" allowfullscreen="true" frameborder="no"></iframe>`))
	}

	return ast.WalkContinue, nil
}
