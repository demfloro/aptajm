package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"

	"gitea.demsh.org/demsh/ircfw"
)

var (
	isUrl       = regexp.MustCompile("https?://[^\\s]+")
	ignoredExts = []string{
		".avi", ".mkv", ".ogg", ".doc", ".docx",
		".xls", ".xlsx", ".mp3", ".flac", ".m3a",
		".torrent", ".png", ".jpg", ".jpeg",
		".gif", ".bmp", ".txt", ".rar", ".zip",
		".gz", ".bz", ".bzip2", ".zstd", ".tgz", ".tar",
	}
)

func isIgnored(ctx context.Context, bot *ircbot, url string) bool {
	var domain, blocked string
	if strings.HasPrefix(url, "https") {
		domain = strings.TrimPrefix(url, "https://")
	} else {
		domain = strings.TrimPrefix(domain, "http://")
	}
	domain = strings.Split(domain, "/")[0]
	err := bot.stmts[ignoredDomain].QueryRowContext(ctx, domain).Scan(&blocked)
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
		if endsWith(url, ignoredExts) || isIgnored(ctx, bot, url) {
			continue
		}
		title, err := getTitle(ctx, url)
		if err != nil {
			bot.Log("Failed to extract title from %q, err: %q", url, err)
			continue
		}
		titles = append(titles, title)
	}

	for _, title := range titles {
		msg.Reply(ctx, []string{fmt.Sprintf("^:: %s", title)})
	}
}

func getTitle(ctx context.Context, url string) (title string, err error) {
	body, err := get(ctx, url, "text/html")
	if err != nil {
		return "", err
	}
	defer body.Close()

	title, err = extractTitle(body)
	if err != nil {
		return "", err
	}
	return title, nil
}

func get(ctx context.Context, url string, contentType string) (body io.ReadCloser, err error) {
	var client http.Client
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	if Config.UserAgent != "" {
		request.Header.Set("User-Agent", Config.UserAgent)
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
	if t := response.Header.Get("Content-Type"); !strings.HasPrefix(t, contentType) {
		response.Body.Close()
		err = fmt.Errorf("Content-Type is not %q: %q", contentType, t)
		return
	}
	body = response.Body
	return
}
