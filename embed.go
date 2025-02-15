package embed

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
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
	EmbededProviderYouTube     = "youtube"
	EmbededProviderBilibili    = "bilibili"
	EmbededProviderTwitter     = "twitter"
	EmbededProviderTradingView = "tradingview"
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

// Embeded struct represents a Embeded embed of the Markdown text.
type Embeded struct {
	ast.Image
	// ast.BaseBlock
	element  string
	Provider string
	VID      string
	Theme    string
}

// KindEmbeded is a NodeKind of the YouTube node.
var KindEmbeded = ast.NewNodeKind("Embeded")

// Kind implements Node.Kind.
func (n *Embeded) Kind() ast.NodeKind {
	return KindEmbeded
}

// NewEmbeded returns a new YouTube node.
func NewEmbeded(c *Embeded) *Embeded {
	c.Destination = c.Image.Destination
	c.Title = c.Image.Title

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
			logrus.WithError(err).Println("goldmark-embed: failed to parse url")
			msg := ast.NewString([]byte(fmt.Sprintf("<!-- %s -->", err)))
			msg.SetCode(true)
			n.Parent().InsertAfter(n.Parent(), n, msg)
			return ast.WalkContinue, nil
		}

		// Embed an object?
		vid := ""
		theme := "dark"
		provider := EmbededProviderYouTube
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
			provider = EmbededProviderBilibili

		} else if u.Host == "twitter.com" || u.Host == "m.twitter.com" || u.Host == "x.com" {
			// https://twitter.com/{username}/status/{id number}?theme=dark
			vid = string(img.Destination)
			if u.Host == "x.com" {
				// replace x.com with twitter.com, because x.com doesn't support using x.com as embeded source
				vid = strings.Replace(vid, "x.com", "twitter.com", 1)
			}
			theme = u.Query().Get("theme")
			provider = EmbededProviderTwitter

		} else if u.Host == "tradingview.com" || u.Host == "www.tradingview.com" {
			// https://www.tradingview.com/chart/UC0wWW9o/?symbol=BITFINEX%3ABTCUSD
			vid = u.Query().Get("symbol")
			theme = u.Query().Get("theme")
			provider = EmbededProviderTradingView

		} else {
			return ast.WalkContinue, nil
		}

		if vid != "" {
			ev := NewEmbeded(
				&Embeded{
					Image:    *img,
					Provider: provider,
					VID:      vid,
					Theme:    theme,
				})
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
	reg.Register(KindEmbeded, r.renderEmbeded)
}

func (r *HTMLRenderer) renderEmbeded(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}

	ev := node.(*Embeded)
	if ev.Provider == EmbededProviderYouTube {
		w.Write([]byte(`<div class="embeded-object-wrapper"><iframe class="embeded-object youtube-embeded-object" width="100%" height="400" src="https://www.youtube.com/embed/` + ev.VID + `" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe></div>`))

	} else if ev.Provider == EmbededProviderBilibili {
		w.Write([]byte(`<div class="embeded-object-wrapper"><iframe class="embeded-object bilibili-embeded-object" width="100%" height="400" src="//player.bilibili.com/player.html?bvid=` + ev.VID + `&page=1" scrolling="no" border="0" framespacing="0" allowfullscreen="true" frameborder="no"></iframe></div>`))

	} else if ev.Provider == EmbededProviderTwitter {
		html, err := GetTweetOembedHtml(ev.VID, ev.Theme)
		if err != nil || html == "" {
			html = fmt.Sprintf(`<div class="embeded-object-wrapper normal-wrapper"><div class="embeded-object twitter-embeded-object normal-object error">Failed to load tweet from %s</div></div>`, ev.VID)
		} else {
			html = fmt.Sprintf(`<div class="embeded-object-wrapper normal-wrapper"><div class="embeded-object twitter-embeded-object normal-object">%s</div></div>`, html)
		}
		w.Write([]byte(html))

	} else if ev.Provider == EmbededProviderTradingView {
		html, err := GetTraddingViewEmbedHtml(ev.VID, ev.Theme)
		if err != nil || html == "" {
			html = fmt.Sprintf(`<div class="embeded-object-wrapper normal-wrapper"><div class="embeded-object tradingview-embeded-object error">Failed to load tradingview chart from %s</div></div>`, ev.VID)
		} else {
			html = fmt.Sprintf(`<div class="embeded-object-wrapper auto-resize"><div class="embeded-object tradingview-embeded-object no-border">%s</div></div>`, html)
		}
		w.Write([]byte(html))
	}

	return ast.WalkContinue, nil
}
