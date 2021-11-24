package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"gitea.demsh.org/demsh/ircfw"
	"gopkg.in/tomb.v2"
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

var (
	ErrExitRequested = errors.New("exit requested")
)

type handler func(ctx context.Context, bot *ircbot, msg ircfw.Msg)

type ircbot struct {
	tomb     *tomb.Tomb
	db       *sql.DB
	client   *ircfw.Client
	handlers map[botCmd]handler
	stmts    map[dbStmt]*sql.Stmt
	logger   ircfw.Logger
	config   ConfigStruct
	mu       sync.Mutex
	// mutex protected fields
	weatherCache  map[string]weather
	currencyCache map[string]string
	bashLimits    map[string]*time.Timer
}

func newIRCBot(baseCtx context.Context, config ConfigStruct, logger ircfw.Logger) (*ircbot, error) {
	t, tombCtx := tomb.WithContext(baseCtx)
	ircbot := ircbot{tomb: t, config: config}
	ircbot.initDB()

	dialer := tls.Dialer{Config: &tls.Config{InsecureSkipVerify: true}}
	ctx, cancel := context.WithTimeout(tombCtx, config.Timeout)
	socket, err := dialer.DialContext(ctx, "tcp", config.Server)
	cancel()
	if err != nil {
		t.Kill(err)
		return nil, err
	}

	handler := func(msg ircfw.Msg) {
		go dispatch(&ircbot, msg)
	}
	client, _ := ircfw.NewClient(
		ircfw.Context(tombCtx),
		ircfw.Nick(config.Nick),
		ircfw.Ident(config.Ident),
		ircfw.RealName(config.Realname),
		ircfw.Password(config.Password),
		ircfw.NickServPass(config.NickservPass),
		ircfw.Socket(socket),
		ircfw.SetLogger(logger),
		ircfw.Handler(handler),
		ircfw.Charmap(config.Charset),
	)
	ircbot.logger = logger
	ircbot.client = client
	ircbot.handlers = initHandlers()

	ircbot.bashLimits = make(map[string]*time.Timer)
	ircbot.tomb.Go(ircbot.finalizer)
	ircbot.tomb.Go(ircbot.pruneWeatherCache)
	ircbot.tomb.Go(ircbot.pruneCurrencyCache)

	for _, channel := range config.Channels {
		ctx, cancel := context.WithTimeout(tombCtx, config.Timeout)
		_, err = ircbot.Join(ctx, channel)
		cancel()
		if err != nil {
			logger.Logf("Error joining channel %q: %q", channel, err)
		}
	}

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

func (b *ircbot) finalizer() error {
	for {
		select {
		case <-b.tomb.Dying():
			b.db.Close()
			b.client.Quit("Bot is dying")
			return tomb.ErrDying
		}
	}
}

func (b *ircbot) Wait() error {
	return b.client.Wait()
}

func (b *ircbot) Quit() {
	b.client.Quit("Normal exit")
	b.tomb.Kill(ErrExitRequested)
}

func (b *ircbot) Log(v ...interface{}) {
	b.logger.Log(v...)
}

func (b *ircbot) Logf(format string, params ...interface{}) {
	b.logger.Logf(format, params...)
}

func (b *ircbot) Debug(format string, params ...interface{}) {
	b.logger.Debug(format, params...)
}

func (b *ircbot) isAdmin(prefix string) bool {
	i := strings.Index(prefix, "!")
	if i == -1 {
		return false
	}
	prefix = prefix[i+1:]
	for _, admin := range b.config.Admins {
		if prefix == admin {
			return true
		}
	}
	return false
}

func handleQuit(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	if !bot.isAdmin(msg.Prefix()) {
		return
	}
	bot.Quit()
}

func handleDefault(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	handleURL(ctx, bot, msg)
}

func dispatch(bot *ircbot, msg ircfw.Msg) {
	text := msg.Text()
	if len(text) < 1 {
		return
	}
	nick := msg.Nick()
	for _, ignore := range bot.config.Ignored {
		if nick == ignore {
			return
		}
	}
	cmd := strings.ToLower(strings.Split(text[0], " ")[0])
	handler, ok := bot.handlers[botCmd(cmd)]
	ctx := bot.tomb.Context(nil)
	ctx, cancel := context.WithTimeout(ctx, bot.config.Timeout)
	bot.Debug("msg: %s", msg)
	if !ok {
		handleDefault(ctx, bot, msg)
		cancel()
		return
	}
	handler(ctx, bot, msg)
	cancel()
}
