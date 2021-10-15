package main

import (
	"fmt"

	"gorm.io/gorm"
)

// Pool to store pool attributes
type Pool struct {
	gorm.Model
	Coin            string `gorm:"unique;not null"`
	LastBlockNumber uint64
}

// NewPool creates a Pool
func NewPool(coin string) *Pool {
	return &Pool{Coin: coin}
}

// String represents Pool to a printable format
func (p *Pool) String() string {
	return fmt.Sprintf("Pool<%s>", p.Coin)
}

// Block to store block attributes
type Block struct {
	Hash   string  `gorm:"unique;not null"`
	Number uint64  `gorm:"not null"`
	Reward float64 `gorm:"not null"`
}

// NewBlock creates a Block
func NewBlock(hash string, number uint64, reward float64) *Block {
	return &Block{
		Hash:   hash,
		Number: number,
		Reward: reward,
	}
}

// String represents Block to a printable format
func (b *Block) String() string {
	return fmt.Sprintf("Block<%d>", b.Number)
}
