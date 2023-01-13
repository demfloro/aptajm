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
	cmdBtc     botCmd = "!btc"
	cmdEth     botCmd = "!eth"
	cmdXmr     botCmd = "!xmr"
	cmdWeather botCmd = "!п"
	cmdStatus  botCmd = "!status"
	cmdQuit    botCmd = "!quit"
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
	config   config
	mu       sync.Mutex
	// mutex protected fields
	weatherCache  map[string]weather
	currencyCache map[string]string
	bashLimits    map[string]*time.Timer
	channels      map[string]*ircfw.Channel
}

func newIRCBot(baseCtx context.Context, conf config, logger ircfw.Logger) (*ircbot, error) {
	t, tombCtx := tomb.WithContext(baseCtx)
	ircbot := ircbot{tomb: t, config: conf}
	ircbot.initDB()

	dialer := tls.Dialer{Config: &tls.Config{InsecureSkipVerify: true}}
	ctx, cancel := context.WithTimeout(tombCtx, conf.timeout)
	socket, err := dialer.DialContext(ctx, "tcp", conf.server)
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
		ircfw.Nick(conf.nick),
		ircfw.Ident(conf.ident),
		ircfw.RealName(conf.realname),
		ircfw.Password(conf.password),
		ircfw.NickServPass(conf.nickservPass),
		ircfw.Socket(socket),
		ircfw.SetLogger(logger),
		ircfw.Handler(handler),
	)
	ircbot.logger = logger
	ircbot.client = client
	ircbot.handlers = initHandlers()

	ircbot.bashLimits = make(map[string]*time.Timer)
	ircbot.channels = make(map[string]*ircfw.Channel)

	ircbot.tomb.Go(ircbot.finalizer)
	ircbot.tomb.Go(ircbot.pruneWeatherCache)
	ircbot.tomb.Go(ircbot.pruneCurrencyCache)
	ircbot.tomb.Go(ircbot.pollNews)

	for _, channel := range conf.channels {
		ctx, cancel := context.WithTimeout(tombCtx, conf.timeout)
		chHandle, err := ircbot.Join(ctx, channel)
		cancel()
		ircbot.channels[chHandle.Name()] = chHandle
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
	<-b.tomb.Dying()
	b.db.Close()
	b.client.Quit("Bot is dying")
	return tomb.ErrDying
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
	for _, admin := range b.config.admins {
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
	for _, ignore := range bot.config.ignored {
		if nick == ignore {
			return
		}
	}
	cmd := strings.ToLower(strings.Split(text[0], " ")[0])
	handler, ok := bot.handlers[botCmd(cmd)]
	ctx := bot.tomb.Context(nil)
	ctx, cancel := context.WithTimeout(ctx, bot.config.timeout)
	bot.Debug("msg: %s", msg)
	if !ok {
		handleDefault(ctx, bot, msg)
		cancel()
		return
	}
	handler(ctx, bot, msg)
	cancel()
}
