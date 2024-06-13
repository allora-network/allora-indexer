package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const eventsRPCQuery = "block_results?height="

type EventType string

const (
	// RewardEvent represents a reward event
	RewardEvent EventType = "reward"
	// ScoreEvent represents a score event
	ScoreEvent EventType = "score"
	// NoneEvent represents an event that doesn't need processing
	NoneEvent EventType = "none"
)

// EventProcessing defines the type of processing needed for an event
type EventProcessing struct {
	Type EventType
}

var event_whitelist = map[string]EventProcessing{
	"emissions.v1.EventScoresSet":      {Type: ScoreEvent},
	"emissions.v1.EventRewardsSettled": {Type: RewardEvent},
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
func FetchEventBlockData(config ClientConfig, height uint64) (*BlockResult, error) {
	url := fmt.Sprintf("%s%s%d", config.Node, eventsRPCQuery, height)
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

// FilterEvents filters events based on a whitelist on its type
func FilterEvents(events []Event, whitelist map[string]EventProcessing) []Event {
	var filteredEvents []Event
	for _, event := range events {
		if processing, ok := whitelist[event.Type]; ok && processing.Type != NoneEvent {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents
}

// processes the events of a block
func processBlock(config ClientConfig, height uint64) error {
	blockData, err := FetchEventBlockData(config, height)
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
