package main

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/tomb.v2"
)

func (b *ircbot) pollNews() error {
	var next time.Duration
	var oldnews string
	now := time.Now()
	min := now.Minute()
	delta := 1 - min
	if delta > 0 {
		next = time.Minute * time.Duration(delta)
	} else if delta == 0 {
		next = time.Minute
	} else {
		next = time.Minute * time.Duration(60+delta)
	}
	rootctx := b.tomb.Context(nil)
	timer := time.NewTimer(next)
	for {
		select {
		case <-b.tomb.Dying():
			timer.Stop()
			return tomb.ErrDying
		case <-timer.C:
			timer.Reset(time.Hour)
			ctx, cancel := context.WithTimeout(rootctx, 10*time.Second)
			line, err := getNewsPost(ctx, "https://t.me/s/neuralmeduza", b.config.userAgent)
			cancel()
			if err != nil {
				b.Logf("Error getting news: %#v", err)
				continue
			}
			if line == oldnews {
				continue
			}
			b.mu.Lock()
			channel := b.channels["#mania"]
			b.mu.Unlock()
			channel.Say(fmt.Sprintf("новости: %s", line))
			oldnews = line
		}
	}
}

func getNewsPost(ctx context.Context, url string, userAgent string) (string, error) {
	body, _, err := get(ctx, url, "text/html", userAgent)
	if err != nil {
		return "", err
	}
	defer body.Close()

	text, err := extractTGLastPost(body)
	if err != nil {
		return "", err
	}
	return text, nil
}
