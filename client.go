package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

// FlexpoolAPIURL constant to store Flexpool API URL
const FlexpoolAPIURL = "https://api.flexpool.io/v2"

// MaxIterations to avoid infinite loop while requesting paged routes on Flexpool API
const MaxIterations = 10

// UserAgent to identify ourselves on the Flexpool API
var UserAgent = fmt.Sprintf("flexassistant/%s", AppVersion)

// FlexpoolClient to store the HTTP client
type FlexpoolClient struct {
	client *http.Client
}

// NewFlexpoolClient to create a client to manage Flexpool API calls
func NewFlexpoolClient() *FlexpoolClient {
	return &FlexpoolClient{
		client: &http.Client{Timeout: time.Second * 3},
	}
}

// request to create an HTTPS request, call the Flexpool API, detect errors and return the result in bytes
func (f *FlexpoolClient) request(url string) ([]byte, error) {
	log.Debugf("Requesting %s", url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", UserAgent)

	resp, err := f.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(jsonBody, &result)

	if result["error"] != nil {
		return nil, fmt.Errorf("Flexpool API error: %s", result["error"].(string))
	}
	return jsonBody, nil
}

// BalanceResponse represents the JSON structure of the Flexpool API response for balance
type BalanceResponse struct {
	Error  string `json:"error"`
	Result struct {
		Balance float64 `json:"balance"`
	} `json:"result"`
}

// MinerBalance returns the current unpaid balance
func (f *FlexpoolClient) MinerBalance(coin string, address string) (float64, error) {
	body, err := f.request(fmt.Sprintf("%s/miner/balance?coin=%s&address=%s", FlexpoolAPIURL, coin, address))
	if err != nil {
		return 0, err
	}

	var response BalanceResponse
	json.Unmarshal(body, &response)
	return response.Result.Balance, nil
}

// PaymentsResponse represents the JSON structure of the Flexpool API response for payments
type PaymentsResponse struct {
	Error  string `json:"error"`
	Result struct {
		Data []struct {
			Hash      string  `json:"hash"`
			Value     float64 `json:"value"`
			Timestamp int64   `json:"timestamp"`
		} `json:"data"`
	} `json:"result"`
}

// MinerPayments returns an ordered list of payments
func (f *FlexpoolClient) MinerPayments(coin string, address string, limit int) (payments []*Payment, err error) {
	page := 0
	for {
		body, err := f.request(fmt.Sprintf("%s/miner/payments/?coin=%s&address=%s&page=%d", FlexpoolAPIURL, coin, address, page))
		if err != nil {
			return nil, err
		}

		var response PaymentsResponse
		json.Unmarshal(body, &response)

		for _, result := range response.Result.Data {
			payment := NewPayment(
				result.Hash,
				result.Value,
				result.Timestamp,
			)
			payments = append(payments, payment)
			if len(payments) >= limit {
				// sort by timestamp
				sort.Slice(payments, func(p1, p2 int) bool {
					return payments[p1].Timestamp > payments[p2].Timestamp
				})
				return payments, nil
			}
		}
		page = page + 1
		if page > MaxIterations {
			return nil, fmt.Errorf("Max iterations of %d reached", MaxIterations)
		}
	}
}

// LastMinerPayment return the last payment of a miner
func (f *FlexpoolClient) LastMinerPayment(miner *Miner) (*Payment, error) {
	log.Debugf("Fetching last payment of %s", miner)
	payments, err := f.MinerPayments(miner.Coin, miner.Address, 1)
	if err != nil {
		return nil, err
	}
	return payments[0], nil
}

// WorkersResponse represents the JSON structure of the Flexpool API response for workers
type WorkersResponse struct {
	Error  string `json:"error"`
	Result []struct {
		Name      string `json:"name"`
		IsOnline  bool   `json:"isOnline"`
		LastSteen int64  `json:"lastSeen"`
	} `json:"result"`
}

// MinerWorkers returns a list of workers given a miner address
func (f *FlexpoolClient) MinerWorkers(coin string, address string) (workers []*Worker, err error) {
	body, err := f.request(fmt.Sprintf("%s/miner/workers?coin=%s&address=%s", FlexpoolAPIURL, coin, address))
	if err != nil {
		return nil, err
	}

	var response WorkersResponse
	json.Unmarshal(body, &response)

	for _, result := range response.Result {
		worker := NewWorker(
			address,
			result.Name,
			result.IsOnline,
			time.Unix(result.LastSteen, 0),
		)
		workers = append(workers, worker)
	}
	return workers, nil
}

// BlocksResponse represents the JSON structure of the Flexpool API response for blocks
type BlocksResponse struct {
	Error  string `json:"error"`
	Result struct {
		Data []struct {
			Hash   string  `json:"hash"`
			Number uint64  `json:"number"`
			Reward float64 `json:"reward"`
		} `json:"data"`
	} `json:"result"`
}

// PoolBlocks returns an ordered list of blocks
func (f *FlexpoolClient) PoolBlocks(coin string, limit int) (blocks []*Block, err error) {
	page := 0
	for {
		body, err := f.request(fmt.Sprintf("%s/pool/blocks/?coin=%s&page=%d", FlexpoolAPIURL, coin, page))
		if err != nil {
			return nil, err
		}

		var response BlocksResponse
		json.Unmarshal(body, &response)

		for _, result := range response.Result.Data {
			block := NewBlock(
				result.Hash,
				result.Number,
				result.Reward,
			)
			blocks = append(blocks, block)
			if len(blocks) >= limit {
				// sort by number
				sort.Slice(blocks, func(b1, b2 int) bool {
					return blocks[b1].Number < blocks[b2].Number
				})
				return blocks, nil
			}
		}
		page = page + 1
		if page > MaxIterations {
			return nil, fmt.Errorf("Max iterations of %d reached", MaxIterations)
		}
	}
}

// LastPoolBlock return the last discovered block for a given pool
func (f *FlexpoolClient) LastPoolBlock(pool *Pool) (*Block, error) {
	blocks, err := f.PoolBlocks(pool.Coin, 1)
	if err != nil {
		return nil, err
	}
	return blocks[0], nil
}

// CoinsResponse represents the JSON structure of the Flexpool API response for pool coins
type CoinsResponse struct {
	Error  string `json:"error"`
	Result struct {
		Coins []struct {
			Ticker string `json:"ticker"`
			Name   string `json:"name"`
		} `json:"coins"`
	} `json:"result"`
}

// RandomPool returns a random pool from the API
func (f *FlexpoolClient) RandomPool() (*Pool, error) {
	log.Debug("Fetching a random pool")
	body, err := f.request(fmt.Sprintf("%s/pool/coins", FlexpoolAPIURL))
	if err != nil {
		return nil, err
	}
	var response CoinsResponse
	json.Unmarshal(body, &response)
	randomIndex := rand.Intn(len(response.Result.Coins))
	if err != nil {
		return nil, err
	}
	randomCoin := response.Result.Coins[randomIndex]
	return NewPool(randomCoin.Ticker), nil
}

// TopMinersResponse represents the JSON structure of the Flexpool API response for pool top miners
type TopMinersResponse struct {
	Error  string `json:"error"`
	Result []struct {
		Address string `json:"address"`
	} `json:"result"`
}

// RandomMiner returns a random miner from the API
func (f *FlexpoolClient) RandomMiner(pool *Pool) (*Miner, error) {
	log.Debug("Fetching a random miner")
	body, err := f.request(fmt.Sprintf("%s/pool/topMiners?coin=%s", FlexpoolAPIURL, pool.Coin))
	if err != nil {
		return nil, err
	}
	var response TopMinersResponse
	json.Unmarshal(body, &response)
	randomResult := response.Result[rand.Intn(len(response.Result))]
	randomMiner, err := NewMiner(randomResult.Address, pool.Coin)
	if err != nil {
		return nil, err
	}
	randomBalance, err := f.MinerBalance(pool.Coin, randomMiner.Address)
	if err != nil {
		return nil, err
	}
	randomMiner.Balance = randomBalance
	return randomMiner, nil
}

// RandomWorker returns a random worker from the API
func (f *FlexpoolClient) RandomWorker(miner *Miner) (*Worker, error) {
	log.Debug("Fetching a random worker")
	workers, err := f.MinerWorkers(miner.Coin, miner.Address)
	if err != nil {
		return nil, err
	}
	return workers[rand.Intn(len(workers))], nil
}
