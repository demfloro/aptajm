package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitea.demsh.org/demsh/ircfw"
)

func (b *ircbot) fetchQuote(ctx context.Context, qid int) (quote, error) {
	var (
		id        int
		timestamp int64
		rating    int
		text      string
		err       error
	)
	switch {
	case qid < 0:
		return quote{}, fmt.Errorf("quote not found")
	case qid > 0:
		err = b.stmts[fetchQuote].QueryRowContext(ctx, qid).Scan(&id, &timestamp, &rating, &text)
	default:
		err = b.stmts[fetchRandomQuote].QueryRowContext(ctx).Scan(&id, &timestamp, &rating, &text)
	}
	if err == sql.ErrNoRows {
		return quote{}, fmt.Errorf("quote not found")
	} else if err != nil {
		return quote{}, err
	}
	return quote{Id: id, Date: time.Unix(timestamp, 0), Rating: rating, Text: strings.Split(text, "\n")}, nil
}

func extractId(lines []string) (int, error) {
	param := strings.Split(lines[0], " ")[0]
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func serveQuote(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	text := removeCmd(msg.Text(), cmdBash)
	id, err := extractId(text)
	if err != nil {
		bot.Log("Failed to parse quote id: %#v", err)
		return
	}
	if id < 0 {
		return
	}
	quote, err := bot.fetchQuote(ctx, id)
	if err != nil {
		bot.Log("Failed to get quote: %#v", err)
		msg.Reply(ctx, []string{fmt.Sprintf("No quote with id %d", id)})
		return
	}
	msg.Reply(ctx, quote.ircFormat())

}

func serveRandomQuote(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	quote, err := bot.fetchQuote(ctx, 0)
	if err != nil {
		bot.Log("Failed to get quote: %#v", err)
		return
	}
	msg.Reply(ctx, quote.ircFormat())
}

func handleBash(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	if !allow(bot, msg) {
		return
	}
	firstline := msg.Text()[0]
	if len(strings.Split(firstline, " ")) != 1 {
		serveQuote(ctx, bot, msg)
		return
	}
	serveRandomQuote(ctx, bot, msg)
}

func allow(bot *ircbot, msg ircfw.Msg) bool {
	if msg.IsPrivate() {
		return true
	}
	channel := msg.Channel().Name()
	bot.mu.Lock()
	limit, ok := bot.bashLimits[channel]
	if !ok {
		bot.Log("Initialised missing bash limiter for %q", channel)
		limit = time.NewTimer(time.Nanosecond)
		bot.bashLimits[channel] = limit
	}
	bot.mu.Unlock()
	select {
	case <-limit.C:
		limit.Reset(time.Minute)
		return true
	default:
	}
	return false
}
