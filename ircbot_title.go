package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"gitea.demsh.org/demsh/ircfw"
)

var (
	isUrl       = regexp.MustCompile(`https?://[^\s]{1,500}`)
	ignoredExts = []string{
		".avi", ".mkv", ".ogg", ".doc", ".docx",
		".xls", ".xlsx", ".mp3", ".flac", ".m3a",
		".torrent", ".png", ".jpg", ".jpeg",
		".gif", ".bmp", ".txt", ".rar", ".zip",
		".gz", ".bz", ".bzip2", ".zstd", ".tgz", ".tar",
	}
)

func isIgnored(ctx context.Context, bot *ircbot, potentialUrl string) bool {
	URL, err := url.Parse(potentialUrl)
	if err != nil {
		return true
	}
	if endsWith(URL.EscapedPath(), ignoredExts) {
		return true
	}
	domain := URL.Hostname()

	var blocked string
	err = bot.stmts[ignoredDomain].QueryRowContext(ctx, domain).Scan(&blocked)
	if err != nil {
		return false
	}
	if domain != blocked {
		return false
	}
	return true
}

func handleURL(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	var titles []string
	text := strings.Join(msg.Text(), " ")
	text = dropRunes(text, "\x01")

	urls := isUrl.FindAllString(text, 2)
	if len(urls) == 0 {
		return
	}
	for _, url := range urls {
		if isIgnored(ctx, bot, url) {
			continue
		}
		title, err := getTitle(ctx, url, bot.config.userAgent)
		if err != nil {
			bot.Log("Failed to extract title from %q, err: %q", url, err)
			continue
		}
		if len(title) == 0 {
			continue
		}
		titles = append(titles, title)
	}

	for _, title := range titles {
		msg.Reply(ctx, []string{fmt.Sprintf("^:: %s", title)})
	}
}

func getTitle(ctx context.Context, url string, userAgent string) (title string, err error) {
	body, utf8, err := get(ctx, url, "text/html", userAgent)
	if err != nil {
		return "", err
	}
	defer body.Close()

	title, err = extractTitle(body, utf8)
	if err != nil {
		return "", err
	}
	return title, nil
}

func get(ctx context.Context, url string, contentType string, userAgent string) (body io.ReadCloser, utf8 bool, err error) {
	var client http.Client
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	if userAgent != "" {
		request.Header.Set("User-Agent", userAgent)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	client.Jar = jar
	response, err := client.Do(request)
	if err != nil {
		return
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		err = fmt.Errorf("wrong status code %d", response.StatusCode)
		return
	}
	responseType := response.Header.Get("Content-Type")
	var ok bool
	if ok, utf8 = checkContentType(responseType, contentType); !ok {
		response.Body.Close()
		err = fmt.Errorf("Content-Type %q is not %q", responseType, contentType)
		return
	}
	body = response.Body
	return
}
