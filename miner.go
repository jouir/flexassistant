package main

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CoinLengthETH represents the size of an ETH address
const CoinLengthETH = 42

// CoinLengthXCH represents the size of an XCH address
const CoinLengthXCH = 62

// Miner to store miner attributes
type Miner struct {
	gorm.Model
	Coin                 string
	Address              string `gorm:"unique;not null"`
	Balance              float64
	LastPaymentTimestamp int64
}

// NewMiner creates a Miner
func NewMiner(address string, coin string) (*Miner, error) {
	miner := &Miner{Address: address, Coin: coin}
	if coin == "" {
		coin, err := miner.ParseCoin()
		if err != nil {
			return nil, err
		}
		miner.Coin = coin
	}
	return miner, nil
}

// ParseCoin deduces the currency given the miner address
func (m *Miner) ParseCoin() (coin string, err error) {
	if m.Address == "" {
		return "", fmt.Errorf("Miner address is empty")
	}
	if len(m.Address) == CoinLengthETH && strings.HasPrefix(m.Address, "0x") {
		return "eth", nil
	}
	if len(m.Address) == CoinLengthXCH && strings.HasPrefix(m.Address, "xch") {
		return "xch", nil
	}
	return "", fmt.Errorf("Unsupported address")
}

// String represents Miner to a printable format
func (m *Miner) String() string {
	return fmt.Sprintf("Miner<%s>", m.Address)
}

// Payment to store payment attributes
type Payment struct {
	Hash      string
	Value     float64
	Timestamp int64
}

// NewPayment creates a Payment
func NewPayment(hash string, value float64, timestamp int64) *Payment {
	return &Payment{
		Hash:      hash,
		Value:     value,
		Timestamp: timestamp,
	}
}

// String represents a Payment to a printable format
func (p *Payment) String() string {
	return fmt.Sprintf("Payment<%s>", p.Hash)
}

// Worker to store workers attributes
type Worker struct {
	gorm.Model
	MinerAddress string    `gorm:"not null"`
	Name         string    `gorm:"not null"`
	IsOnline     bool      `gorm:"not null"`
	LastSeen     time.Time `gorm:"not null"`
}

// NewWorker creates a Worker
func NewWorker(minerAddress string, name string, isOnline bool, lastSeen time.Time) *Worker {
	return &Worker{
		MinerAddress: minerAddress,
		Name:         name,
		IsOnline:     isOnline,
		LastSeen:     lastSeen,
	}
}

// String represents Worker to a printable format
func (w *Worker) String() string {
	return fmt.Sprintf("Worker<%s>", w.Name)
}
