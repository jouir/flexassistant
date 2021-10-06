package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// request to create an HTTPS request, call the Flexpool API, detect errors and return the result
func (f *FlexpoolClient) request(url string) (result map[string]interface{}, err error) {
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

	json.Unmarshal(jsonBody, &result)

	if result["error"] == nil {
		return result["result"].(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("Flexpool API error: %s", result["error"].(string))
}

// MinerBalance returns the current unpaid balance
func (f *FlexpoolClient) MinerBalance(coin string, address string) (float64, error) {
	response, err := f.request(fmt.Sprintf("%s/miner/balance?coin=%s&address=%s", FlexpoolAPIURL, coin, address))
	if err != nil {
		return 0, err
	}
	return response["balance"].(float64), nil
}

// MinerPayments returns an ordered list of payments
func (f *FlexpoolClient) MinerPayments(coin string, address string, limit int) (payments []*Payment, err error) {
	page := 0
	for {
		response, err := f.request(fmt.Sprintf("%s/miner/payments/?coin=%s&address=%s&page=%d", FlexpoolAPIURL, coin, address, page))
		if err != nil {
			return nil, err
		}

		for _, result := range response["data"].([]interface{}) {
			raw := result.(map[string]interface{})
			payment := NewPayment(
				raw["hash"].(string),
				raw["value"].(float64),
				raw["timestamp"].(float64),
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

// PoolBlocks returns an ordered list of blocks
func (f *FlexpoolClient) PoolBlocks(coin string, limit int) (blocks []*Block, err error) {
	page := 0
	for {
		response, err := f.request(fmt.Sprintf("%s/pool/blocks/?coin=%s&page=%d", FlexpoolAPIURL, coin, page))
		if err != nil {
			return nil, err
		}

		for _, result := range response["data"].([]interface{}) {
			raw := result.(map[string]interface{})
			block := NewBlock(
				raw["hash"].(string),
				raw["number"].(float64),
				raw["reward"].(float64),
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
