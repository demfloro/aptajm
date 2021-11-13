package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"gitea.demsh.org/demsh/ircfw"
	"golang.org/x/text/encoding/charmap"
)

type dbStmt uint8
type botCmd string

const (
	cmdBash    botCmd = "!bash"
	cmdBtc            = "!btc"
	cmdEth            = "!eth"
	cmdXmr            = "!xmr"
	cmdWeather        = "!п"
	cmdStatus         = "!status"
	cmdQuit           = "!quit"
)

const (
	unknown dbStmt = iota
	fetchQuote
	fetchRandomQuote
	fetchRandomRating
	fetchCity
	ignoredDomain
)

type handler func(ctx context.Context, bot *ircbot, msg ircfw.Msg)

type ircbot struct {
	db       *sql.DB
	client   *ircfw.Client
	handlers map[botCmd]handler
	stmts    map[dbStmt]*sql.Stmt
	stop     context.CancelFunc
	logger   *log.Logger
	mu       sync.Mutex
	// mutex protected fields
	weatherCache  map[string]weather
	currencyCache map[string]string
	bashLimits    map[string]*time.Timer
}

func newIRCBot(baseCtx context.Context, dbname string, nick string,
	ident string, realName string, password string, nickservPass string,
	proto string, server string,
	charmap *charmap.Charmap, logger *log.Logger) (*ircbot, error) {

	db, cancelDB, err := initDB(baseCtx, dbname, logger)
	if err != nil {
		return nil, err
	}

	dialer := tls.Dialer{Config: &tls.Config{InsecureSkipVerify: true}}
	ctx, cancel := context.WithTimeout(baseCtx, Config.Timeout)
	socket, err := dialer.DialContext(ctx, proto, server)
	cancel()
	if err != nil {
		return nil, err
	}

	ircbot := ircbot{db: db}
	handler := func(msg ircfw.Msg) {
		go dispatch(baseCtx, &ircbot, msg)
	}

	client, cancelClient := ircfw.NewClient(baseCtx, nick, ident,
		realName, password, nickservPass, socket, logger, handler, charmap)

	ircbot.logger = logger
	ircbot.client = client
	ircbot.handlers = initHandlers()

	ctx, cancel = context.WithTimeout(baseCtx, Config.Timeout)
	ircbot.stmts, err = initStmts(ctx, db)
	cancel()
	if err != nil {
		return nil, err
	}
	ircbot.weatherCache = make(map[string]weather)
	ircbot.currencyCache = make(map[string]string)
	ircbot.bashLimits = make(map[string]*time.Timer)
	go pruneWeatherCache(baseCtx, &ircbot)
	go pruneCurrencyCache(baseCtx, &ircbot)

	stop := func() {
		ircbot.mu.Lock()
		for _, limit := range ircbot.bashLimits {
			limit.Stop()
		}
		ircbot.mu.Unlock()
		cancelClient()
		cancelDB()
	}
	ircbot.stop = stop
	return &ircbot, nil
}

func initHandlers() map[botCmd]handler {
	return map[botCmd]handler{
		cmdBash:    handleBash,
		cmdWeather: handleWeather,
		cmdBtc:     handleCurrencies,
		cmdEth:     handleCurrencies,
		cmdXmr:     handleCurrencies,
		cmdStatus:  handleStatus,
		cmdQuit:    handleQuit,
	}

}

func (b *ircbot) Join(ctx context.Context, channel string) (*ircfw.Channel, error) {
	ch, err := b.client.Join(ctx, channel)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.bashLimits[channel] = time.NewTimer(time.Nanosecond)
	b.mu.Unlock()
	return ch, nil
}

func (b *ircbot) Wait() {
	b.client.Wait()
}

func (b *ircbot) Quit() {
	b.client.Quit("Normal exit")
	b.db.Close()
	b.stop()
}

func (b *ircbot) Log(format string, params ...interface{}) {
	b.logger.Printf(format, params...)
}

func isAdmin(prefix string) bool {
	i := strings.Index(prefix, "!")
	if i == -1 {
		return false
	}
	prefix = prefix[i+1:]
	for _, admin := range Config.Admins {
		if prefix == admin {
			return true
		}
	}
	return false
}

func handleQuit(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	if !isAdmin(msg.Prefix()) {
		return
	}
	bot.Quit()
}

func handleDefault(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	handleURL(ctx, bot, msg)
}

func dispatch(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	text := msg.Text()
	if len(text) < 1 {
		return
	}
	nick := msg.Nick()
	for _, ignore := range Config.Ignored {
		if nick == ignore {
			return
		}
	}
	cmd := strings.ToLower(strings.Split(text[0], " ")[0])
	handler, ok := bot.handlers[botCmd(cmd)]
	ctx, cancel := context.WithTimeout(ctx, Config.Timeout)
	bot.Log("msg: %s", msg)
	if !ok {
		handleDefault(ctx, bot, msg)
		cancel()
		return
	}
	handler(ctx, bot, msg)
	cancel()
}
