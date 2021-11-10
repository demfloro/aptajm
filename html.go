package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
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

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node, depth uint) (string, bool) {
	depth++
	if depth == RecursionLimit {
		return "", false
	}
	if isTitleElement(n) && n.FirstChild != nil {
		return n.FirstChild.Data, true
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		result, ok := traverse(child, depth)
		if ok {
			return result, ok
		}
	}
	return "", false
}

func extractTitle(data io.Reader) (title string, err error) {
	data, err = htmlToUTF8(data)
	if err != nil {
		return
	}
	tree, err := html.Parse(data)
	if err != nil {
		return
	}
	title, ok := traverse(tree, 0)
	if !ok {
		return "", errors.New("failed to find title")
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
	return
}
