package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"gitea.demsh.org/demsh/ircfw"
	"gopkg.in/tomb.v2"
)

const (
	weatherURL = "https://api.openweathermap.org/data/2.5/weather?units=metric&lang=ru&APPID=%s&q=%s"
	mmHgMagic  = 0.750062
)

type weather struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity int     `json:"humidity"`
		TempMin  float64 `json:"temp_min"`
		TempMax  float64 `json:"temp_max"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Sys struct {
		Country string `json:"country"`
	} `json:"sys"`
	CityID  int    `json:"id"`
	Name    string `json:"name"`
	Cod     int    `json:"cod"`
	created time.Time
}

func (w weather) expired() bool {
	if time.Since(w.created) > 10*time.Minute {
		return true
	}
	return false
}

func (w weather) String() string {
	return fmt.Sprintf("%s/%s: %s; температура: %.1f °C; давление: %.1f мм рт.ст; ветер %.1f м/с, относ. влаж.: %d%%",
		w.Name, w.Sys.Country, w.Weather[0].Description,
		w.Main.Temp, w.Main.Pressure*mmHgMagic, w.Wind.Speed,
		w.Main.Humidity)
}

func handleWeather(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	params := strings.Split(removeCmd(msg.Text(), cmdWeather)[0], " ")
	if len(params) < 1 {
		return
	}
	city := strings.ToLower(params[0])
	weather, err := getWeather(ctx, bot, city)
	if err != nil {
		bot.Log("Weather for %q: %q", city, err)
		return
	}
	msg.Reply(ctx, []string{weather.String()})
}

func getWeather(ctx context.Context, bot *ircbot, alias string) (result weather, err error) {
	var (
		city, country string
	)
	err = bot.stmts[fetchCity].QueryRowContext(ctx, alias).Scan(&city, &country)
	if err != nil {
		return
	}
	city = fmt.Sprintf("%s,%s", city, country)
	bot.mu.Lock()
	result, ok := bot.weatherCache[city]
	bot.mu.Unlock()
	if ok {
		return
	}
	body, _, err := get(ctx, fmt.Sprintf(weatherURL, bot.config.WeatherToken, city), "application/json", bot.config.UserAgent)
	if err != nil {
		return
	}
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return
	}
	bot.mu.Lock()
	bot.weatherCache[city] = result
	bot.mu.Unlock()
	return
}

func (b *ircbot) pruneWeatherCache() error {
	b.weatherCache = make(map[string]weather)
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-b.tomb.Dying():
			ticker.Stop()
			return tomb.ErrDying
		case <-ticker.C:
		}
		b.mu.Lock()
		for city, weather := range b.weatherCache {
			if weather.expired() {
				delete(b.weatherCache, city)
			}
		}
		b.mu.Unlock()
	}
}
