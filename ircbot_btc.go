package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"gitea.demsh.org/demsh/ircfw"
	"gopkg.in/tomb.v2"
)

const (
	btcURL = "https://min-api.cryptocompare.com/data/pricemulti?fsyms=%s&tsyms=USD"
)

func handleCurrencies(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	currency := strings.ToLower(msg.Text()[0][1:])
	bot.mu.Lock()
	price, ok := bot.currencyCache[currency]
	bot.mu.Unlock()
	if ok {
		msg.Reply(ctx, []string{fmt.Sprintf("%s/USD: %s", strings.ToUpper(currency), price)})
		return
	}
	price, err := getPrice(ctx, currency, bot.config.userAgent)
	if err != nil {
		bot.Log("Failed to get price for %q: %q", currency, err)
		return
	}
	bot.mu.Lock()
	bot.currencyCache[currency] = price
	bot.mu.Unlock()
	msg.Reply(ctx, []string{fmt.Sprintf("%s/USD: %s", strings.ToUpper(currency), price)})
}

func getPrice(ctx context.Context, currency string, userAgent string) (price string, err error) {
	body, _, err := get(ctx, fmt.Sprintf(btcURL, currency), "application/json", userAgent)
	if err != nil {
		return "", err
	}
	defer body.Close()

	return extractPrice(currency, body)
}

func extractPrice(currency string, data io.Reader) (price string, err error) {
	var raw interface{}
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return "", err
	}
	array := raw.(map[string]interface{})
	switch name := array[strings.ToUpper(currency)].(type) {
	case map[string]interface{}:
		switch maybePrice := name["USD"].(type) {
		case float64:
			price = fmt.Sprintf("%.2f", maybePrice)
			return
		default:
			return "", fmt.Errorf("Error parsing: %q", b)
		}
	default:
		return "", fmt.Errorf("Error parsing: %q", b)
	}
}

func (b *ircbot) pruneCurrencyCache() error {
	b.currencyCache = make(map[string]string)
	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-b.tomb.Dying():
			ticker.Stop()
			return tomb.ErrDying
		case <-ticker.C:
		}

		b.mu.Lock()
		if len(b.currencyCache) != 0 {
			b.currencyCache = make(map[string]string)
		}
		b.mu.Unlock()
	}
}
