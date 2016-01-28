package ticketswap

import (
	"errors"
	"fmt"

	"time"

	"github.com/tucnak/telebot"
)

var aboutText = fmt.Sprint(`They call me the Ticketbot, I can help you find tickets at ticketswap.com.

You can control me by sending these commands:
/help - get the help info
/startwatch - start the ads monitoring
/stopwatch - stop the ads monitoring`)

// Bot represents a separate ticketswap bot instance.
type Bot struct {
	Token     string
	IsRunning bool
	telebot   *telebot.Bot
	messages  chan telebot.Message
	tasks     map[string]*Task
}

// NewBot creates a Bots with token `token`, which is a secret API key assigned to particular bot.
func NewBot(token string) (*Bot, error) {
	bot, err := telebot.NewBot(token)
	if err != nil {
		return nil, errors.New("Failed to create a bot!")
	}

	return &Bot{token, false, bot, make(chan telebot.Message), make(map[string]*Task)}, nil
}

// Start periodically pulls messages from the Telegram chat, checks for
// available ads at the ticketswap.com and sends a message if any.
// The timeout `timeout` defines the timeout to use while sending requests to the ticketswap.
func (bot *Bot) Start(timeout time.Duration) {
	bot.telebot.Listen(bot.messages, 1*time.Second)
	for message := range bot.messages {
		cmd, err := NewCommand(message.Text)
		if err != nil {
			bot.telebot.SendMessage(message.Chat, fmt.Sprint("The command is not valid! Use /help to start."), nil)
			continue
		}

		switch cmd.CommandType {
		case tCmdHelp:
			bot.telebot.SendMessage(message.Chat, aboutText, &telebot.SendOptions{DisableWebPagePreview: true})
		case tCmdStartWatch:
			taskID := message.Chat.Destination() + cmd.Argv[0]
			if _, ok := bot.tasks[taskID]; ok {
				bot.telebot.SendMessage(message.Chat, fmt.Sprint("I monitor the ads for this url already! Use /stopwatch command to stop monitoring."), nil)
				continue
			}

			task := NewTask(cmd.Argv[0])
			bot.tasks[taskID] = task
			go func(adsCh <-chan Advertisements, bot *Bot, chat telebot.Chat) {
				for ads := range adsCh {
					bot.telebot.SendMessage(chat, fmt.Sprintf(
						"Found the next ads:\n %s", ads.String()), &telebot.SendOptions{DisableWebPagePreview: true})
				}
				fmt.Println("end of ads")
			}(task.AdsCh, bot, message.Chat)
			task.Start(timeout)
			bot.telebot.SendMessage(message.Chat, fmt.Sprint(
				"Started the ads monitoring! Will notify you when the ticket appears. Use /stopwatch command to stop me."), nil)
			fmt.Printf("started the ads monitoring %s\n", task.URL)
		case tCmdStopWatch:
			taskID := message.Chat.Destination() + cmd.Argv[0]
			task, ok := bot.tasks[taskID]
			if !ok {
				bot.telebot.SendMessage(message.Chat, fmt.Sprint("I don't monitor the ads for this url! Use /startwatch command to start monitoring."), nil)
				continue
			}

			task.Stop()
			delete(bot.tasks, taskID)
			bot.telebot.SendMessage(message.Chat, fmt.Sprint(
				"Stopped the ads monitoring! Use /startwatch command to start monitoring again."), nil)
			fmt.Printf("stopped the ads monitoring %s\n", task.URL)
		}
	}
}

func (bot *Bot) Stop() {
	bot.IsRunning = false
	for _, task := range bot.tasks {
		task.Stop()
	}
}
