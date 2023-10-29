# goldmark-embed

This is a fork from https://github.com/13rac1/goldmark-embed, to use Markdown `![]()` image embed syntax to support additional object formats.

[goldmark]: http://github.com/yuin/goldmark

## Supported Providers

* YouTube Video
* Bilibili Video
* X's Tweet Oembed Widget
* TradingView Widget

## Demo

This markdown:

```md
![](https://youtu.be/dQw4w9WgXcQ?si=0kalBBLQpIXT1Wcd)
```

```md
![](https://www.bilibili.com/video/BV1uT4y1P7CX)
```

```md
![](https://twitter.com/NASA/status/1704954156149084293)
```

```md
![](https://www.tradingview.com/chart/AA0aBB8c/?symbol=BITFINEX%3ABTCUSD)
```

### Installation

```bash
go get github.com/quail.ink/goldmark-embed
```

## Usage

```go
  markdown := goldmark.New(
    goldmark.WithExtensions(
      embed.Embed,
    ),
  )
  var buf bytes.Buffer
  if err := markdown.Convert([]byte(source), &buf); err != nil {
    panic(err)
  }
  fmt.Print(buf)
}
```

## TODO

* Embed Options
* Additional Data Sources

## License

MIT

## Author

Brad Erickson
