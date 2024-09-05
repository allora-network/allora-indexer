package main

import (
	"strings"

	"github.com/allora-network/allora-indexer/types"
	"github.com/rs/zerolog/log"
)

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
