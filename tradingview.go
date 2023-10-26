package embed

import (
	"bytes"
	"fmt"
	"math/rand"
	"text/template"
)

const tpl = `
<!-- TradingView Widget BEGIN -->
<div class="tradingview-widget-container" style="height:100%;width:100%">
  <div id="{{.ID}}" style="height:calc(100% - 32px);width:100%"></div>
  <div class="tradingview-widget-copyright"><a href="https://www.tradingview.com/" rel="noopener nofollow" target="_blank"><span class="blue-text">Track all markets on TradingView</span></a></div>
  <script type="application/javascript" src="https://s3.tradingview.com/tv.js"></script>
  <script type="application/javascript">
  new TradingView.widget(
  {
		"autosize": true,
		"symbol": "{{.Symbol}}",
		"interval": "D",
		"timezone": "Etc/UTC",
		"theme": "{{.Theme}}",
		"style": "1",
		"locale": "en",
		"enable_publishing": false,
		"allow_symbol_change": true,
		"container_id": "{{.ID}}"
	}
  );
  </script>
</div>
<!-- TradingView Widget END -->
`

func GetTraddingViewEmbedHtml(symbol, theme string) (string, error) {
	if theme == "dark" {
		theme = "dark"
	} else {
		theme = "light"
	}

	id := fmt.Sprintf("tradingview_%d", rand.Intn(1000))

	// compile the template
	t, err := template.New("tradingview").Parse(tpl)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	if err = t.Execute(&buf, map[string]string{
		"ID":     id,
		"Symbol": symbol,
		"Theme":  theme,
	}); err != nil {
		return "", err
	}

	// html := fmt.Sprintf(tpl, id, symbol, theme, id)

	return buf.String(), nil
}
