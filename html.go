package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	RecursionLimit = 1000
)

/*
 * Kudos to https://siongui.github.io/
 * https://siongui.github.io/2018/10/27/auto-detect-and-convert-html-encoding-to-utf8-in-go/
 * https://siongui.github.io/2016/05/10/go-get-html-title-via-net-html/
 */

type filterFunc func(n *html.Node) bool
type extractFunc func(n *html.Node) (string, bool)

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func defaultExtractor(n *html.Node) (string, bool) {
	if n.FirstChild != nil {
		return strings.TrimSpace(n.FirstChild.Data), true
	}
	return "", false
}

func twitchExtractor(n *html.Node) (string, bool) {
	if n == nil {
		return "", false
	}
	for _, attr := range n.Attr {
		if attr.Key == "content" {
			return strings.TrimSpace(attr.Val), true
		}
	}
	return "", false

}

func tgExtractor(n *html.Node) (string, bool) {
	if n == nil {
		return "", false
	}
	return n.FirstChild.Data, true
}

func isTGElement(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "tgme_widget_message_text js-message_text" {
				return true
			}
		}
	}
	return false
}

func extractTGLastPost(data io.Reader) (quote string, err error) {
	tree, err := html.Parse(data)
	if err != nil {
		return
	}
	quote, ok := revTraverse(tree, 0, isTGElement, tgExtractor)
	if !ok {
		return "", errors.New("failed to extract post")
	}
	return
}

func isTwitchElement(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "meta" {
		for _, attr := range n.Attr {
			if attr.Key == "property" && attr.Val == "og:title" {
				return true
			}
		}
	}
	return false
}

func traverse(n *html.Node, depth uint, filter filterFunc, extractor extractFunc) (string, bool) {
	depth++
	if depth == RecursionLimit {
		return "", false
	}
	if filter(n) {
		return extractor(n)
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		result, ok := traverse(child, depth, filter, extractor)
		if ok {
			return result, ok
		}
	}
	return "", false
}

func revTraverse(n *html.Node, depth uint, filter filterFunc, extractor extractFunc) (string, bool) {
	depth++
	if depth == RecursionLimit {
		return "", false
	}
	if filter(n) {
		return extractor(n)
	}

	for child := n.LastChild; child != nil; child = child.PrevSibling {
		result, ok := revTraverse(child, depth, filter, extractor)
		if ok {
			return result, ok
		}
	}
	return "", false
}

func extractTitle(data io.Reader, utf8 bool) (title string, err error) {
	if !utf8 {
		data, err = htmlToUTF8(data)
		if err != nil {
			return
		}
	}
	tree, err := html.Parse(data)
	if err != nil {
		return
	}
	title, ok := traverse(tree, 0, isTitleElement, defaultExtractor)
	if !ok {
		return "", errors.New("failed to find title")
	}
	switch title {
	case "Twitch":
		title, ok = traverse(tree, 0, isTwitchElement, twitchExtractor)
		if !ok {
			title = "Twitch"
		}
	}
	return
}

func htmlToUTF8(data io.Reader) (result io.Reader, err error) {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}
	encoding, _, _, err := determineEncoding(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	result = transform.NewReader(bytes.NewReader(b), encoding.NewDecoder())
	return

}
func determineEncoding(data io.Reader) (e encoding.Encoding, name string, certain bool, err error) {
	b, err := bufio.NewReader(data).Peek(1024)
	if err != nil {
		return
	}
	e, name, certain = charset.DetermineEncoding(b, "")
	/* hack for websites like interfax.ru which don't conform to
	 * HTML standard and put their Content-Type tag beyond 1024 bytes
	 * https://html.spec.whatwg.org/multipage/parsing.html#determining-the-character-encoding
	 */
	if e == charmap.Windows1252 && !certain {
		e = charmap.Windows1251
	}
	return
}

func checkContentType(header string, contentType string) (ok bool, utf8 bool) {
	header = strings.TrimSpace(header)
	values := strings.Split(header, ";")
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		switch value {
		case contentType:
			ok = true
		case "charset=utf-8":
			utf8 = true
		}
	}
	return
}
