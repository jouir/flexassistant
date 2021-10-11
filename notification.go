package main

import (
	"fmt"
	"strings"

	"github.com/leekchan/accounting"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

// Notifier interface to define how to send all kind of notifications
type Notifier interface {
	NotifyBalance(miner Miner, difference float64) error
	NotifyPayment(miner Miner, payment Payment) error
	NotifyBlock(pool Pool, block Block) error
	NotifyOfflineWorker(worker Worker) error
}

// TelegramNotifier to send notifications using Telegram
// Implements the Notifier interface
type TelegramNotifier struct {
	bot         *telegram.BotAPI
	chatID      int64
	channelName string
}

// NewTelegramNotifier to create a TelegramNotifier
func NewTelegramNotifier(config *TelegramConfig) (*TelegramNotifier, error) {
	bot, err := telegram.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	log.Debugf("Connected to Telegram as %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:         bot,
		chatID:      config.ChatID,
		channelName: config.ChannelName,
	}, nil
}

// sendMessage to send a generic message on Telegram
func (t *TelegramNotifier) sendMessage(message string) error {
	var request telegram.MessageConfig
	if t.chatID != 0 {
		request = telegram.NewMessage(t.chatID, message)
	} else {
		request = telegram.NewMessageToChannel(t.channelName, message)
	}
	request.DisableWebPagePreview = true
	request.ParseMode = telegram.ModeMarkdown

	response, err := t.bot.Send(request)
	if err != nil {
		return err
	}
	log.Debugf("Message %d sent to Telegram", response.MessageID)
	return nil
}

// NotifyBalance to format and send a notification when the unpaid balance has changed
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyBalance(miner Miner) error {
	ac := accounting.Accounting{
		Symbol:    strings.ToUpper(miner.Coin),
		Precision: 6,
		Format:    "%v %s",
	}
	convertedBalance, err := ConvertCurrency(miner.Coin, miner.Balance)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("💰 *Balance* _%s_", ac.FormatMoney(convertedBalance))
	return t.sendMessage(message)
}

// NotifyPayment to format and send a notification when a new payment has been detected
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyPayment(miner Miner, payment Payment) error {
	ac := accounting.Accounting{
		Symbol:    strings.ToUpper(miner.Coin),
		Precision: 6,
		Format:    "%v %s",
	}
	convertedValue, err := ConvertCurrency(miner.Coin, payment.Value)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("💵 *Payment* _%s_", ac.FormatMoney(convertedValue))
	return t.sendMessage(message)
}

// NotifyBlock to format and send a notification when a new block has been detected
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyBlock(pool Pool, block Block) error {
	precision := 6
	if pool.Coin == "xch" {
		precision = 2
	}
	ac := accounting.Accounting{
		Symbol:    strings.ToUpper(pool.Coin),
		Precision: precision,
		Format:    "%v %s",
	}

	convertedValue, err := ConvertCurrency(pool.Coin, block.Reward)
	if err != nil {
		return err
	}
	verb, err := ConvertAction(pool.Coin)
	if err != nil {
		return err
	}
	url, err := FormatBlockURL(pool.Coin, block.Hash)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("🎉 *%s* [#%d](%s) _%s_", verb, block.Number, url, ac.FormatMoney(convertedValue))
	return t.sendMessage(message)
}

// NotifyOfflineWorker sends a message when a worker is online or offline
func (t *TelegramNotifier) NotifyOfflineWorker(worker Worker) error {
	stateIcon := "🟢"
	stateMessage := "online"
	if !worker.IsOnline {
		stateIcon = "🔴"
		stateMessage = "offline"
	}
	message := fmt.Sprintf("%s *Worker* _%s_ is %s", stateIcon, worker.Name, stateMessage)
	return t.sendMessage(message)
}
