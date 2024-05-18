package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/allora-network/allora-cosmos-pump/types"
	"github.com/rs/zerolog/log"
)


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
		block, err := fetchBlock(config, height)
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetchBlock block height")
			return
		}
		writeBlock(config, block)
		lastProcessedHeight = latestHeight
	}

}

func fetchBlock(config ClientConfig, height int64) (types.BlockQuery, error){
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
		return types.BlockQuery{}, err
	}

	var blockQuery types.BlockQuery
	if err := json.Unmarshal(output, &blockQuery); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal block info")
		return types.BlockQuery{}, err
	}

	return blockQuery, nil
	// Process the block information (e.g., insert into database)
	// processBlockQuery(config, blockQuery)
}

func getLatestBlockHeightFromDB() (int64, error) {
	// Use sql.NullInt64 which can handle NULL values
	var maxHeight sql.NullInt64
	err := dbPool.QueryRow(context.Background(), "SELECT MAX(height) FROM block_info").Scan(&maxHeight)
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

func writeBlock(config ClientConfig, blockQuery types.BlockQuery) {
	// Process the block information (e.g., insert into database)
	// Assuming `insertBlockInfo` is defined elsewhere
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

	err = insertEncTxs(height, blockQuery.Data.Txs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert Txs")
	}

}
