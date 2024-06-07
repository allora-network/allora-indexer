package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://allora-rpc.edgenet.allora.network/block_results?height="

var event_whitelist = map[string]bool{
	"inferer_rewards_settled":               true,
	"forecaster_rewards_settled":            true,
	"reputer_and_delegator_rewards_settled": true,
	"inferer_scores_set":                    true,
	"reputer_scores_set":                    true,
	"forecaster_scores_set":                 true,
}

type BlockResult struct {
	Result struct {
		Height              string  `json:"height"`
		FinalizeBlockEvents []Event `json:"finalize_block_events"`
	} `json:"result"`
}

type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FetchEventBlockData fetches block data for a given height
func FetchEventBlockData(height uint64) (*BlockResult, error) {
	url := fmt.Sprintf("%s%d", baseURL, height)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var blockResult BlockResult
	err = json.Unmarshal(body, &blockResult)
	if err != nil {
		return nil, err
	}

	return &blockResult, nil
}

func FilterEvents(events []Event, whitelist map[string]bool) []Event {
	var filteredEvents []Event
	for _, event := range events {
		if whitelist[event.Type] {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents
}

func processBlock(db *sql.DB, height uint64) error {
	blockData, err := FetchEventBlockData(height)
	if err != nil {
		return fmt.Errorf("failed to fetch block data: %w", err)
	}

	filteredEvents := FilterEvents(blockData.Result.FinalizeBlockEvents, event_whitelist)

	var eventRecords []EventRecord
	for _, event := range filteredEvents {
		data, err := json.Marshal(event.Attributes)
		if err != nil {
			return fmt.Errorf("failed to marshal event attributes: %w", err)
		}

		var sender string
		for _, attr := range event.Attributes {
			if attr.Key == "sender" {
				sender = attr.Value
				break
			}
		}

		eventRecords = append(eventRecords, EventRecord{
			Height: height,
			Type:   event.Type,
			Sender: sender,
			Data:   data,
		})
	}

	err = insertEvents(eventRecords)
	if err != nil {
		return fmt.Errorf("failed to insert events: %w", err)
	}

	return nil
}
