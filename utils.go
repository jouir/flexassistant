package main

import (
	"fmt"
)

// WeisToETHDivider to divide Weis to ETH
const WeisToETHDivider = 1000000000000000000

// MojoToXCHDivider to divide Mojo to XCH
const MojoToXCHDivider = 1000000000000

// ConvertCurrency divides the smallest unit of the currency to the currency itself
// Example: for "eth", convert from Weis to ETH
func ConvertCurrency(coin string, value float64) (float64, error) {
	switch coin {
	case "etc":
		return ConvertWeis(value), nil
	case "eth":
		return ConvertWeis(value), nil
	case "xch":
		return ConvertMojo(value), nil
	default:
		return 0, fmt.Errorf("Coin %s not supported", coin)
	}
}

// ConvertWeis converts the value from Weis to ETH
func ConvertWeis(value float64) float64 {
	return value / WeisToETHDivider
}

// ConvertMojo converts the value from Mojo to XCH
func ConvertMojo(value float64) float64 {
	return value / MojoToXCHDivider
}

// FormatBlockURL returns the URL on the respective blockchain explorer given the coin and the block hash
func FormatBlockURL(coin string, hash string) (string, error) {
	switch coin {
	case "etc":
		return fmt.Sprintf("https://etcblockexplorer.com/block/%s", hash), nil
	case "eth":
		return fmt.Sprintf("https://etherscan.io/block/%s", hash), nil
	case "xch":
		return fmt.Sprintf("https://www.chiaexplorer.com/blockchain/block/%s", hash), nil
	}
	return "", fmt.Errorf("Coin %s not supported", coin)
}

// FormatTransactionURL returns the URL on the respective blockchain explorer given the coin and the transaction hash
func FormatTransactionURL(coin string, hash string) (string, error) {
	switch coin {
	case "etc":
		return fmt.Sprintf("https://etcblockexplorer.com/address/%s", hash), nil
	case "eth":
		return fmt.Sprintf("https://etherscan.io/tx/%s", hash), nil
	case "xch":
		return fmt.Sprintf("https://www.chiaexplorer.com/blockchain/coin/%s", hash), nil
	}
	return "", fmt.Errorf("Coin %s not supported", coin)
}
