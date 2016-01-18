package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"

	tswp "github.com/sgl0v/go-tswp"
)

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func getTickets(ticketsUrl string) tswp.Tickets {
	response, err := http.Get(ticketsUrl)
	perror(err)

	// 2. parse html
	defer response.Body.Close()
	tickets := tswp.NewTickets(response.Body)
	sort.Sort(tickets)
	return tickets
}

func getAds(url string, adsCh chan<- tswp.Advertisements) {
	// 1. fetch html
	response, _ := http.Get(url)

	// 2. parse html
	defer response.Body.Close()
	ads := tswp.NewAdvertisements(response.Body)
	sort.Sort(ads)
	adsCh <- ads
}

func sendToTelegram(token string, chatId int, msg string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	body := []byte(fmt.Sprintf(`{"chat_id": %d, "text": "%s"}`, chatId, msg))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	perror(err)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	perror(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Failed to send the telegram message!"))
	}
}

func messageSender(msgCh <-chan string, exitCh chan<- bool) {
	for msg := range msgCh {
		sendToTelegram(tBotToken, tChatId, msg)
	}
	exitCh <- true
}

const (
	tBotToken = "143596761:AAFuA-bjKYygRKXWFPZwNWHTghAAjXueVAg"
	tChatId   = 133119219
	adUrl     = "https://www.ticketswap.com/event/rihanna/floor/c1671553-db2b-4f0f-b9c1-51a70e6b48e0/4857"
)

func main() {
	adsCh := make(chan tswp.Advertisements)
	exitCh := make(chan bool)
	msgCh := make(chan string)
	go getAds(adUrl, adsCh)
	go messageSender(msgCh, exitCh)

	for {
		select {
		case ads := <-adsCh:
			if ads != nil {
				msgCh <- ads.String()
			}
			close(msgCh)
		case <-exitCh:
			return
		}
	}
}
