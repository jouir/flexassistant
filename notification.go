package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

//go:embed templates
var templateFiles embed.FS

// Attachment is used to attach objects to templates
type Attachment struct {
	Miner   Miner
	Payment Payment
	Pool    Pool
	Block   Block
	Worker  Worker
}

// Notifier interface to define how to send all kind of notifications
type Notifier interface {
	NotifyBalance(miner Miner, difference float64) error
	NotifyPayment(miner Miner, payment Payment) error
	NotifyBlock(pool Pool, block Block) error
	NotifyOfflineWorker(worker Worker) error
	NotifyTest(client FlexpoolClient) error
}

// TelegramNotifier to send notifications using Telegram
// Implements the Notifier interface
type TelegramNotifier struct {
	bot            *telegram.BotAPI
	chatID         int64
	channelName    string
	configurations *NotificationsConfig
}

// NewTelegramNotifier to create a TelegramNotifier
func NewTelegramNotifier(config *TelegramConfig, configurations *NotificationsConfig) (*TelegramNotifier, error) {
	bot, err := telegram.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	log.Debugf("Connected to Telegram as %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:            bot,
		chatID:         config.ChatID,
		channelName:    config.ChannelName,
		configurations: configurations,
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

// formatMessage to create a message with a template file name (either embeded or on disk)
func (t *TelegramNotifier) formatMessage(templateFileName string, attachment interface{}) (message string, err error) {
	// Create template
	templateName := path.Base(templateFileName)
	templateFunctions := template.FuncMap{
		"upper":                strings.ToUpper,
		"lower":                strings.ToLower,
		"convertCurrency":      ConvertCurrency,
		"formatBlockURL":       FormatBlockURL,
		"formatTransactionURL": FormatTransactionURL,
	}
	tmpl := template.New(templateName).Funcs(templateFunctions)

	// Parse template
	if fileExists(templateFileName) {
		log.Debugf("Parsing template file %s", templateFileName)
		tmpl, err = tmpl.ParseFiles(templateFileName)
	} else {
		log.Debugf("Parsing embeded template file %s", templateFileName)
		tmpl, err = tmpl.ParseFS(templateFiles, templateFileName)
	}
	if err != nil {
		return "", fmt.Errorf("parse failed: %v", err)
	}

	// Execute template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, attachment)
	if err != nil {
		return "", fmt.Errorf("execute failed: %v", err)
	}

	// Extract and return the formatted message
	message = buffer.String()
	return message, nil
}

// isFileTemplate returns true when the filename is a real file on the filesystem
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

// NotifyBalance to format and send a notification when the unpaid balance has changed
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyBalance(miner Miner) (err error) {
	templateName := "templates/balance.tmpl"
	if t.configurations.Balance.Template != "" {
		templateName = t.configurations.Balance.Template
	}
	message, err := t.formatMessage(templateName, Attachment{Miner: miner})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// testNotifyBalance sends a fake balance notification
func (t *TelegramNotifier) testNotifyBalance(client FlexpoolClient) error {
	log.Debug("Testing balance notification")
	randomPool, err := client.RandomPool()
	if err != nil {
		return err
	}
	randomMiner, err := client.RandomMiner(randomPool)
	if err != nil {
		return err
	}
	return t.NotifyBalance(*randomMiner)
}

// NotifyPayment to format and send a notification when a new payment has been detected
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyPayment(miner Miner, payment Payment) error {
	templateName := "templates/payment.tmpl"
	if t.configurations.Payment.Template != "" {
		templateName = t.configurations.Payment.Template
	}
	message, err := t.formatMessage(templateName, Attachment{Miner: miner, Payment: payment})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// testNotifyPayment sends a fake payment notification
func (t *TelegramNotifier) testNotifyPayment(client FlexpoolClient) error {
	log.Debug("Testing payment notification")
	randomPool, err := client.RandomPool()
	if err != nil {
		return err
	}
	randomMiner, err := client.RandomMiner(randomPool)
	if err != nil {
		return err
	}
	randomPayment, err := client.LastMinerPayment(randomMiner)
	if err != nil {
		return err
	}
	return t.NotifyPayment(*randomMiner, *randomPayment)
}

// NotifyBlock to format and send a notification when a new block has been detected
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyBlock(pool Pool, block Block) error {
	templateName := "templates/block.tmpl"
	if t.configurations.Block.Template != "" {
		templateName = t.configurations.Block.Template
	}
	message, err := t.formatMessage(templateName, Attachment{Pool: pool, Block: block})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// testNotifyBlock sends a random block notification
func (t *TelegramNotifier) testNotifyBlock(client FlexpoolClient) error {
	log.Debug("Testing block notification")
	randomPool, err := client.RandomPool()
	if err != nil {
		return err
	}
	randomBlock, err := client.LastPoolBlock(randomPool)
	if err != nil {
		return err
	}
	return t.NotifyBlock(*randomPool, *randomBlock)
}

// NotifyOfflineWorker sends a message when a worker is online or offline
func (t *TelegramNotifier) NotifyOfflineWorker(worker Worker) error {
	templateName := "templates/offline-worker.tmpl"
	if t.configurations.OfflineWorker.Template != "" {
		templateName = t.configurations.OfflineWorker.Template
	}
	message, err := t.formatMessage(templateName, Attachment{Worker: worker})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// testNotifyOfflineWorker sends a fake worker offline notification
func (t *TelegramNotifier) testNotifyOfflineWorker(client FlexpoolClient) error {
	log.Debug("Testing offline worker notification")
	randomBlock, err := client.RandomPool()
	if err != nil {
		return err
	}
	randomMiner, err := client.RandomMiner(randomBlock)
	if err != nil {
		return err
	}
	randomWorker, err := client.RandomWorker(randomMiner)
	if err != nil {
		return err
	}
	log.Debugf("%s", randomWorker)
	return t.NotifyOfflineWorker(*randomWorker)
}

// NotifyTest sends fake notifications
func (t *TelegramNotifier) NotifyTest(client FlexpoolClient) (executed bool, err error) {
	if t.configurations.Balance.Test {
		if err = t.testNotifyBalance(client); err != nil {
			return false, err
		} else {
			executed = true
		}
	}

	if t.configurations.Payment.Test {
		if err = t.testNotifyPayment(client); err != nil {
			return false, err
		} else {
			executed = true
		}
	}

	if t.configurations.Block.Test {
		if err = t.testNotifyBlock(client); err != nil {
			return false, err
		} else {
			executed = true
		}
	}

	if t.configurations.OfflineWorker.Test {
		if err = t.testNotifyOfflineWorker(client); err != nil {
			return false, err
		} else {
			executed = true
		}
	}
	return executed, nil
}
