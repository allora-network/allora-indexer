package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type DBConsensusParams struct {
	MaxBytes         string
	MaxGas           string
	MaxAgeDuration   string
	MaxAgeNumBlocks  string
	EvidenceMaxBytes string
	PubKeyTypes      string // This can be a JSON-encoded array or a comma-separated list
}

type DBBlockInfo struct {
	BlockHash                  string
	BlockTotalParts            int
	BlockPartSetHeaderHash     string
	BlockVersion               string
	ChainID                    string
	Height                     string
	BlockTime                  time.Time
	LastBlockHash              string
	LastBlockTotalParts        int
	LastBlockPartSetHeaderHash string
	LastCommitHash             string
	DataHash                   string
	ValidatorsHash             string
	NextValidatorsHash         string
	ConsensusHash              string
	AppHash                    string
	LastResultsHash            string
	EvidenceHash               string
	ProposerAddress            string
}

var dbConn *pgx.Conn

func initDB(dataSourceName string) {
	var err error
	dbConn, err = pgx.Connect(context.Background(), dataSourceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	setupDB()
}

func setupDB() {
	executeSQL(createBlockInfoTableSQL())
	executeSQL(createConsensusParamsTableSQL())
}

func executeSQL(sqlStatement string) {
	if _, err := dbConn.Exec(context.Background(), sqlStatement); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute SQL statement: %v\n", err)
		os.Exit(1)
	}
}

func createBlockInfoTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS block_info (
		block_hash VARCHAR(255),
		block_total_parts INT,
		block_part_set_header_hash VARCHAR(255),
		block_version VARCHAR(255),
		chain_id VARCHAR(255),
		height VARCHAR(255) PRIMARY KEY,
		block_time TIMESTAMP,
		last_block_hash VARCHAR(255),
		last_block_total_parts INT,
		last_block_part_set_header_hash VARCHAR(255),
		last_commit_hash VARCHAR(255),
		data_hash VARCHAR(255),
		validators_hash VARCHAR(255),
		next_validators_hash VARCHAR(255),
		consensus_hash VARCHAR(255),
		app_hash VARCHAR(255),
		last_results_hash VARCHAR(255),
		evidence_hash VARCHAR(255),
		proposer_address VARCHAR(255)
	);`
}

func createConsensusParamsTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS consensus_params (
		id SERIAL PRIMARY KEY,
		max_bytes VARCHAR(255),
		max_gas VARCHAR(255),
		max_age_duration VARCHAR(255),
		max_age_num_blocks VARCHAR(255),
		evidence_max_bytes VARCHAR(255),
		pub_key_types TEXT
	);`
}

func insertBlockInfo(blockInfo DBBlockInfo) error {
	_, err := dbConn.Exec(context.Background(), `
		INSERT INTO block_info (
			block_hash, 
			block_total_parts,
			block_part_set_header_hash,
			block_version,
			chain_id,
			height,
			block_time,
			last_block_hash,
			last_block_total_parts,
			last_block_part_set_header_hash,
			last_commit_hash,
			data_hash,
			validators_hash,
			next_validators_hash,
			consensus_hash,
			app_hash,
			last_results_hash,
			evidence_hash,
			proposer_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		blockInfo.BlockHash, blockInfo.BlockTotalParts, blockInfo.BlockPartSetHeaderHash,
		blockInfo.BlockVersion, blockInfo.ChainID, blockInfo.Height, blockInfo.BlockTime,
		blockInfo.LastBlockHash, blockInfo.LastBlockTotalParts, blockInfo.LastBlockPartSetHeaderHash,
		blockInfo.LastCommitHash, blockInfo.DataHash, blockInfo.ValidatorsHash,
		blockInfo.NextValidatorsHash, blockInfo.ConsensusHash, blockInfo.AppHash,
		blockInfo.LastResultsHash, blockInfo.EvidenceHash, blockInfo.ProposerAddress,
	)
	if err != nil {
		// Check if the error is due to a unique constraint violation
		if isUniqueViolation(err) {
			log.Info().Msgf("Block height %d already exists in the database. Skipping insert.", blockInfo.Height)
			return nil // or return an error if you prefer
		}
		// Handle other types of errors
		return err
	}
	return nil
}

func insertConsensusParams(params DBConsensusParams) error {
	_, err := dbConn.Exec(context.Background(), `
        INSERT INTO consensus_params (
            max_bytes,
            max_gas,
            max_age_duration,
            max_age_num_blocks,
            evidence_max_bytes,
            pub_key_types
        ) VALUES ($1, $2, $3, $4, $5, $6)`,
		params.MaxBytes,
		params.MaxGas,
		params.MaxAgeDuration,
		params.MaxAgeNumBlocks,
		params.EvidenceMaxBytes,
		params.PubKeyTypes,
	)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}
	return nil
}

func closeDB() {
	if dbConn != nil {
		dbConn.Close(context.Background())
	}
}

func isUniqueViolation(err error) bool {
	// This function depends on your database driver
	// For example, with PostgreSQL using pq driver:
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // 23505 is the code for unique violation in PostgreSQL
	}
	return false
}
