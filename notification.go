package main

import (
	"bytes"
	"embed"
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
}

// TelegramNotifier to send notifications using Telegram
// Implements the Notifier interface
type TelegramNotifier struct {
	bot             *telegram.BotAPI
	chatID          int64
	channelName     string
	templatesConfig *NotificationTemplatesConfig
}

// NewTelegramNotifier to create a TelegramNotifier
func NewTelegramNotifier(config *TelegramConfig, templatesConfig *NotificationTemplatesConfig) (*TelegramNotifier, error) {
	bot, err := telegram.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	log.Debugf("Connected to Telegram as %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:             bot,
		chatID:          config.ChatID,
		channelName:     config.ChannelName,
		templatesConfig: templatesConfig,
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
	// Deduce if template file is embeded or is a file on disk
	embeded := true
	if _, err = os.Stat(templateFileName); os.IsExist(err) {
		embeded = false
	}
	// Reinitialize the error because it was only used for the test
	err = nil

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
	if embeded {
		log.Debugf("Parsing embeded template file %s", templateFileName)
		tmpl, err = tmpl.ParseFS(templateFiles, templateFileName)
	} else {
		log.Debugf("Parsing template file %s", templateFileName)
		tmpl, err = tmpl.ParseFiles(templateFileName)
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

// NotifyBalance to format and send a notification when the unpaid balance has changed
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyBalance(miner Miner) (err error) {
	templateName := "templates/balance.tmpl"
	if t.templatesConfig.Balance != "" {
		templateName = t.templatesConfig.Balance
	}
	message, err := t.formatMessage(templateName, Attachment{Miner: miner})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// NotifyPayment to format and send a notification when a new payment has been detected
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyPayment(miner Miner, payment Payment) error {
	templateName := "templates/payment.tmpl"
	if t.templatesConfig.Payment != "" {
		templateName = t.templatesConfig.Payment
	}
	message, err := t.formatMessage(templateName, Attachment{Miner: miner, Payment: payment})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// NotifyBlock to format and send a notification when a new block has been detected
// Implements the Notifier interface
func (t *TelegramNotifier) NotifyBlock(pool Pool, block Block) error {
	templateName := "templates/block.tmpl"
	if t.templatesConfig.Block != "" {
		templateName = t.templatesConfig.Block
	}
	message, err := t.formatMessage(templateName, Attachment{Pool: pool, Block: block})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}

// NotifyOfflineWorker sends a message when a worker is online or offline
func (t *TelegramNotifier) NotifyOfflineWorker(worker Worker) error {
	templateName := "templates/offline-worker.tmpl"
	if t.templatesConfig.OfflineWorker != "" {
		templateName = t.templatesConfig.OfflineWorker
	}
	message, err := t.formatMessage(templateName, Attachment{Worker: worker})
	if err != nil {
		return err
	}
	return t.sendMessage(message)
}
