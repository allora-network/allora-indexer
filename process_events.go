package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

const eventsRPCQuery = "/block_results?height="

type EventType string

const (
	// RewardEvent represents a reward event
	RewardEvent EventType = "reward"
	// ScoreEvent represents a score event
	ScoreEvent EventType = "score"
	// NetworkLossEvent represents a network loss event
	NetworkLossEvent EventType = "networkloss"
	// ForecastTaskScoreEvent represents a forecast task score event
	ForecastTaskScoreEvent EventType = "forecastTaskScore"
	// ActorLastCommitEvent represents a last commit event
	ActorLastCommitEvent EventType = "lastcommit"
	// TopicRewardEvent represents a topic reward event
	TopicRewardEvent EventType = "topicReward"
	// EMAScoreEvent represents a ema score event
	EMAScoreEvent EventType = "emascore"
	// NoneEvent represents an event that doesn't need processing
	NoneEvent EventType = "none"
	// an invalid event type
	InvalidType EventType = "invalid"
)

// EventProcessing defines the type of processing needed for an event
type EventProcessing struct {
	Type EventType
}

var event_whitelist = map[string]EventProcessing{
	"EventScoresSet":            {Type: ScoreEvent},
	"EventRewardsSettled":       {Type: RewardEvent},
	"EventNetworkLossSet":       {Type: NetworkLossEvent},
	"EventForecastTaskScoreSet": {Type: ForecastTaskScoreEvent},
	"EventWorkerLastCommitSet":  {Type: ActorLastCommitEvent},
	"EventReputerLastCommitSet": {Type: ActorLastCommitEvent},
	"EventTopicRewardsSet":      {Type: TopicRewardEvent},
	"EventEMAScoresSet":         {Type: EMAScoreEvent},
}

type BlockResult struct {
	Result struct {
		Height              string    `json:"height"`
		FinalizeBlockEvents []Event   `json:"finalize_block_events"`
		TxsBlockEvents      []TxEvent `json:"txs_results"`
	} `json:"result"`
}

type TxEvent struct {
	Code       int     `json:"code"`
	Data       string  `json:"data"`
	Log        string  `json:"log"`
	Info       string  `json:"info"`
	Gas_wanted string  `json:"gas_wanted"`
	Gas_used   string  `json:"gas_used"`
	Events     []Event `json:"events"`
	Codespace  string  `json:"codespace"`
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
func FilterEvents(events *BlockResult, whitelist map[string]EventProcessing) []Event {
	var filteredEvents []Event
	for _, event := range events.Result.FinalizeBlockEvents {
		baseType := getBaseEventType(event.Type)
		if baseType == string(InvalidType) {
			log.Debug().Str("type", event.Type).Msg("event type %s is invalid")

			continue
		}
		if processing, ok := whitelist[baseType]; ok && processing.Type != NoneEvent {
			filteredEvents = append(filteredEvents, event)
		}
	}
	for _, blockevent := range events.Result.TxsBlockEvents {
		for _, event := range blockevent.Events {
			baseType := getBaseEventType(event.Type)
			if baseType == string(InvalidType) {
				log.Debug().Str("type", event.Type).Msg("ERROR: event type %s is invalid")
				continue
			}
			if processing, ok := whitelist[baseType]; ok && processing.Type != NoneEvent {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}
	return filteredEvents
}

// getBaseEventType extracts the base event type without prefix
func getBaseEventType(eventType string) string {
	parts := strings.Split(eventType, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1] // Return the last part, e.g., "EventScoresSet"
	}
	return string(InvalidType) // Return InvalidType for invalid types
}

// processes the events of a block
func processBlock(config ClientConfig, height uint64) error {
	blockData, err := FetchEventBlockData(config, height)
	if err != nil {
		return fmt.Errorf("failed to fetch block data: %w", err)
	}
	filteredEvents := FilterEvents(blockData, event_whitelist)

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
