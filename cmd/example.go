package main

import (
	"time"

	tswp "github.com/sgl0v/go-ticketswap"
)

func main() {
	bot, err := tswp.NewBot(TELEGRAM_BOT_TOCKEN)
	if err != nil {
		return
	}
	bot.Start(time.Duration(120))
	defer func() {
		bot.Stop()
	}()
}
