package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AppName to store application name
var AppName string = "flexassistant"

// AppVersion to set version at compilation time
var AppVersion string = "9999"

// GitCommit to set git commit at compilation time (can be empty)
var GitCommit string

// GoVersion to set Go version at compilation time
var GoVersion string

// MaxPayments defaults
const MaxPayments = 10

// MaxBlocks defaults
const MaxBlocks = 50

// initialize logging
func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	config := NewConfig()
	version := flag.Bool("version", false, "Print version and exit")
	quiet := flag.Bool("quiet", false, "Log errors only")
	verbose := flag.Bool("verbose", false, "Print more logs")
	debug := flag.Bool("debug", false, "Print even more logs")
	configFileName := flag.String("config", AppName+".yaml", "Configuration file name")
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	// Logs and configuration
	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *verbose {
		log.SetLevel(log.InfoLevel)
	}
	if *quiet {
		log.SetLevel(log.ErrorLevel)
	}

	if *configFileName != "" {
		err := config.ReadFile(*configFileName)
		if err != nil {
			log.Fatalf("Cannot parse configuration file: %v", err)
		}
	}

	// Database
	var db *gorm.DB
	db, err := NewDatabase(config.DatabaseFile)
	if err != nil {
		log.Fatalf("Could not create database: %v", err)
	}

	if err := CreateDatabaseObjects(db); err != nil {
		log.Fatalf("Could not create objects: %v", err)
	}

	// API client
	client := NewFlexpoolClient()

	// Notifications
	notifier, err := NewTelegramNotifier(&config.TelegramConfig)
	if err != nil {
		log.Fatalf("Could not create notifier: %v", err)
	}

	// Limits
	var maxPayments int
	if config.MaxPayments > 0 {
		maxPayments = config.MaxPayments
	} else {
		maxPayments = MaxPayments
	}

	var maxBlocks int
	if config.MaxBlocks > 0 {
		maxBlocks = config.MaxBlocks
	} else {
		maxBlocks = MaxBlocks
	}

	// Handle miners
	for _, configuredMiner := range config.Miners {
		miner, err := NewMiner(configuredMiner.Address)
		if err != nil {
			log.Warnf("Could not parse miner: %v", err)
			continue
		}

		var dbMiner Miner
		trx := db.Where(Miner{Address: miner.Address}).Attrs(Miner{Address: miner.Address, Coin: miner.Coin}).FirstOrCreate(&dbMiner)
		if trx.Error != nil {
			log.Warnf("Cannot fetch miner %s from database: %v", miner, trx.Error)
		}

		// Balance management
		if configuredMiner.EnableBalance {
			// Balance have never been persisted, skip notifications
			notify := true
			if dbMiner.Balance == 0 {
				notify = false
			}

			log.Debugf("Fetching balance for %s", miner)
			balance, err := client.MinerBalance(miner.Coin, miner.Address)
			if err != nil {
				log.Warnf("Could not fetch unpaid balance: %v", err)
				continue
			}
			log.Debugf("Unpaid balance %.0f", balance)
			miner.Balance = balance
			if miner.Balance != dbMiner.Balance {
				dbMiner.Balance = balance
				if trx = db.Save(&dbMiner); trx.Error != nil {
					log.Warnf("Cannot update miner: %v", trx.Error)
					continue
				}
				if notify {
					err = notifier.NotifyBalance(*miner)
					if err != nil {
						log.Warnf("Cannot send notification: %v", err)
						continue
					}
					log.Infof("Balance notification sent for %s", miner)
				}
			}
		}

		// Payments management
		if configuredMiner.EnablePayments {
			// Payments have never been persisted, skip notifications
			notify := true
			if dbMiner.LastPaymentTimestamp == 0 {
				notify = false
			}

			log.Debugf("Fetching payments for %s", miner)

			payments, err := client.MinerPayments(miner.Coin, miner.Address, maxPayments)
			if err != nil {
				log.Warnf("Could not fetch payments: %v", err)
				continue
			}
			for _, payment := range payments {
				log.Debugf("Fetched %s", payment)
				if dbMiner.LastPaymentTimestamp < payment.Timestamp {
					dbMiner.LastPaymentTimestamp = payment.Timestamp
					if trx = db.Save(&dbMiner); trx.Error != nil {
						log.Warnf("Cannot update miner: %v", trx.Error)
						continue
					}
					if notify {
						err = notifier.NotifyPayment(*miner, *payment)
						if err != nil {
							log.Warnf("Cannot send notification: %v", err)
							continue
						}
						log.Infof("Payment notification sent for %s", payment)
					}
				}
			}
		}
	}

	// Handle pools
	for _, configuredPool := range config.Pools {

		pool := NewPool(configuredPool.Coin)

		var dbPool Pool
		trx := db.Where(Pool{Coin: pool.Coin}).Attrs(Pool{Coin: pool.Coin}).FirstOrCreate(&dbPool)
		if trx.Error != nil {
			log.Warnf("Cannot fetch pool %s from database: %v", pool, trx.Error)
		}

		// Blocks management
		if configuredPool.EnableBlocks {

			// Block number has never been persisted, skip notifications
			notify := true
			if dbPool.LastBlockNumber == 0 {
				notify = false
			}

			log.Debugf("Fetching blocks for %s", pool)
			blocks, err := client.PoolBlocks(pool.Coin, maxBlocks)
			if err != nil {
				log.Warnf("Could not fetch blocks: %v", err)
			} else {
				for _, block := range blocks {
					log.Debugf("Fetched %s", block)
					if dbPool.LastBlockNumber < block.Number {
						dbPool.LastBlockNumber = block.Number
						if trx = db.Save(&dbPool); trx.Error != nil {
							log.Warnf("Cannot update pool: %v", trx.Error)
							continue
						}
						if notify {
							err = notifier.NotifyBlock(*pool, *block)
							if err != nil {
								log.Warnf("Cannot send notification: %v", err)
								continue
							}
							log.Infof("Block notification sent for %s", block)
						}
					}
				}
			}
		}
	}
}

func showVersion() {
	if GitCommit != "" {
		AppVersion = fmt.Sprintf("%s-%s", AppVersion, GitCommit)
	}
	fmt.Printf("%s version %s (compiled with %s)\n", AppName, AppVersion, GoVersion)
}
