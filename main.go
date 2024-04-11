package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/allora-network/allora-cosmos-pump/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var lastProcessedHeight int64 = 0

type Command struct {
	Parts []string
}

type ClientConfig struct {
	Node     string
	CliApp   string
	Commands map[string]Command
}

func ExecuteCommand(cliApp, node string, parts []string) ([]byte, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("no command parts provided")
	}

	var completeParts []string
	for _, part := range parts {
		completeParts = append(completeParts, part)
	}

	completeParts = replacePlaceholders(completeParts, "{node}", node)
	completeParts = replacePlaceholders(completeParts, "{cliApp}", cliApp)

	log.Info().Strs("command", completeParts).Msg("Executing command")
	cmd := exec.Command(completeParts[0], completeParts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Command execution failed")
		return nil, err
	}

	return output, nil
}

func replacePlaceholders(parts []string, placeholder, value string) []string {
	var replacedParts []string
	for _, part := range parts {
		if part == placeholder {
			replacedParts = append(replacedParts, value)
		} else {
			replacedParts = append(replacedParts, part)
		}
	}
	return replacedParts
}

func ExecuteCommandByKey[T any](config ClientConfig, key string) (T, error) {
	var result T

	cmd, ok := config.Commands[key]
	if !ok {
		return result, fmt.Errorf("command not found")
	}

	log.Info().Str("commandName", key).Msg("Starting execution")
	output, err := ExecuteCommand(config.CliApp, config.Node, cmd.Parts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		return result, err
	}

	log.Debug().Str("raw output", string(output)).Msg("Command raw output")

	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Error().Err(err).Str("json", string(output)).Msg("Failed to unmarshal JSON")
		return result, err
	}

	return result, nil
}

func main() {
	var (
		nodeFlag       string
		cliAppFlag     string
		connectionFlag string
	)

	flag.StringVar(&nodeFlag, "node", "https://default-node-address:443", "Node address")
	flag.StringVar(&cliAppFlag, "cliApp", "simd", "CLI app to execute commands")
	flag.StringVar(&connectionFlag, "conn", "postgres://default:password@localhost:5432/database", "Database connection string")
	flag.Parse()

	initDB(connectionFlag)
	defer closeDB()

	// Initialize the lastProcessedHeight with the latest block height from the database
	var err error
	lastProcessedHeight, err = getLatestBlockHeightFromDB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get the latest block height from the database")
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// define the commands to execute payloads
	config := ClientConfig{
		Node:   nodeFlag,
		CliApp: cliAppFlag,
		Commands: map[string]Command{
			"latestBlock": {
				Parts: []string{"{cliApp}", "query", "consensus", "comet", "block-latest", "--node", "{node}", "--output", "json"},
			},
			"blockByHeight": { // Add a template command for fetching blocks by height
				Parts: []string{"{cliApp}", "query", "block", "{height}", "--type=height", "--node", "{node}", "--output", "json"},
			},
			"consensusParams": {
				Parts: []string{"{cliApp}", "query", "consensus", "params", "--node", "{node}", "--output", "json"},
			},
			"decodeTx": {
				Parts: []string{"{cliApp}", "tx", "decode", "params", "--node", "{node}", "--output", "json"},
			},
		},
	}

	// get the latest consensus params in the block
	processConsensusParams(config)

	// Fetch and process the latest block once before starting the loop
	processLatestBlock(config)

	// Set up a ticker to check for new blocks every 4 seconds
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	// Set up a channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			processLatestBlock(config)
		case <-signalChan:
			log.Info().Msg("Shutdown signal received, exiting...")
			return
		}
	}
}

func processConsensusParams(config ClientConfig) {
	consensusParams, err := ExecuteCommandByKey[types.ConsensusParams](config, "consensusParams")
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		return
	}

	err = insertConsensusParams(DBConsensusParams{
		MaxBytes:         consensusParams.Params.Block.MaxBytes,
		MaxGas:           consensusParams.Params.Block.MaxGas,
		MaxAgeDuration:   consensusParams.Params.Evidence.MaxAgeDuration,
		MaxAgeNumBlocks:  consensusParams.Params.Evidence.MaxAgeNumBlocks,
		EvidenceMaxBytes: consensusParams.Params.Evidence.MaxBytes,
		PubKeyTypes:      strings.Join(consensusParams.Params.Validator.PubKeyTypes, ","),
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		return
	}
}

func processLatestBlock(config ClientConfig) {
	blockInfo, err := ExecuteCommandByKey[types.BlockInfo](config, "latestBlock")
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch the latest block")
		return
	}

	latestHeight, err := strconv.ParseInt(blockInfo.Block.Header.Height, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse latest block height")
		return
	}

	for height := lastProcessedHeight + 1; height < latestHeight; height++ {
		fetchAndProcessBlock(config, height)
	}

	processBlock(blockInfo)
	lastProcessedHeight = latestHeight
}

func fetchAndProcessBlock(config ClientConfig, height int64) {
	// Convert height to string
	heightStr := strconv.FormatInt(height, 10)

	// Clone the original command and replace {height} placeholder
	blockCommand := make([]string, len(config.Commands["blockByHeight"].Parts))
	copy(blockCommand, config.Commands["blockByHeight"].Parts)
	for i, part := range blockCommand {
		if part == "{height}" {
			blockCommand[i] = heightStr
		}
	}

	// Execute the command with the updated height
	log.Info().Str("commandName", "blockByHeight").Msgf("Fetching block at height %s", heightStr)
	output, err := ExecuteCommand(config.CliApp, config.Node, blockCommand)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to fetch block at height %s", heightStr)
		return
	}

	var blockQuery types.BlockQuery
	if err := json.Unmarshal(output, &blockQuery); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal block info")
		return
	}

	// Process the block information (e.g., insert into database)
	processBlockQuery(blockQuery)
}

func getLatestBlockHeightFromDB() (int64, error) {
	// Use sql.NullInt64 which can handle NULL values
	var maxHeight sql.NullInt64
	err := dbConn.QueryRow(context.Background(), "SELECT MAX(height) FROM block_info").Scan(&maxHeight)
	if err != nil {
		return 0, fmt.Errorf("failed to query the latest block height: %v", err)
	}

	// Check if maxHeight is valid (not NULL)
	if !maxHeight.Valid {
		// No valid maxHeight found, probably because there are no entries in the table
		return 0, nil // Returning 0 is safe if you treat it as "start from the beginning"
	}

	return maxHeight.Int64, nil
}

func processBlockQuery(blockQuery types.BlockQuery) {
	height, err := strconv.ParseInt(blockQuery.Header.Height, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse block height")
		return
	}

	err = insertBlockInfo(DBBlockInfo{
		BlockHash:                  blockQuery.Header.LastBlockID.Hash,
		BlockTime:                  blockQuery.Header.Time,
		BlockVersion:               blockQuery.Header.Version.Block,
		ChainID:                    blockQuery.Header.ChainID,
		Height:                     height,
		LastBlockHash:              blockQuery.Header.LastBlockID.Hash,
		LastBlockTotalParts:        blockQuery.Header.LastBlockID.PartSetHeader.Total,
		LastBlockPartSetHeaderHash: blockQuery.Header.LastBlockID.PartSetHeader.Hash,
		LastCommitHash:             blockQuery.Header.LastCommitHash,
		DataHash:                   blockQuery.Header.DataHash,
		ValidatorsHash:             blockQuery.Header.ValidatorsHash,
		NextValidatorsHash:         blockQuery.Header.NextValidatorsHash,
		ConsensusHash:              blockQuery.Header.ConsensusHash,
		AppHash:                    blockQuery.Header.AppHash,
		LastResultsHash:            blockQuery.Header.LastResultsHash,
		EvidenceHash:               blockQuery.Header.EvidenceHash,
		ProposerAddress:            blockQuery.Header.ProposerAddress,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to insert block info")
	}

}

func processBlock(blockInfo types.BlockInfo) {
	// Process the block information (e.g., insert into database)
	// Assuming `insertBlockInfo` is defined elsewhere
	height, err := strconv.ParseInt(blockInfo.Block.Header.Height, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse block height")
		return
	}

	err = insertBlockInfo(DBBlockInfo{
		BlockHash:                  blockInfo.Block.Header.LastBlockID.Hash,
		BlockTime:                  blockInfo.Block.Header.Time,
		BlockTotalParts:            blockInfo.BlockID.PartSetHeader.Total,
		BlockPartSetHeaderHash:     blockInfo.BlockID.PartSetHeader.Hash,
		BlockVersion:               blockInfo.Block.Header.Version.Block,
		ChainID:                    blockInfo.Block.Header.ChainID,
		Height:                     height,
		LastBlockHash:              blockInfo.Block.Header.LastBlockID.Hash,
		LastBlockTotalParts:        blockInfo.Block.Header.LastBlockID.PartSetHeader.Total,
		LastBlockPartSetHeaderHash: blockInfo.Block.Header.LastBlockID.PartSetHeader.Hash,
		LastCommitHash:             blockInfo.Block.Header.LastCommitHash,
		DataHash:                   blockInfo.Block.Header.DataHash,
		ValidatorsHash:             blockInfo.Block.Header.ValidatorsHash,
		NextValidatorsHash:         blockInfo.Block.Header.NextValidatorsHash,
		ConsensusHash:              blockInfo.Block.Header.ConsensusHash,
		AppHash:                    blockInfo.Block.Header.AppHash,
		LastResultsHash:            blockInfo.Block.Header.LastResultsHash,
		EvidenceHash:               blockInfo.Block.Header.EvidenceHash,
		ProposerAddress:            blockInfo.Block.Header.ProposerAddress,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to insert block info")
	}
}
