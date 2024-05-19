package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Height                     uint64
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

var dbPool *pgxpool.Pool //*pgx.Conn

func initDB(dataSourceName string) {
	var err error
	// dbPool, err = pgx.Connect(context.Background(), dataSourceName)

	dbConfig, err := pgxpool.ParseConfig(dataSourceName)
	if err!=nil {
	 log.Fatal().Err(err).Msg("Failed to create a config, error: ")
	}
	dbPool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	setupDB()
}

func closeDB() {
	if dbPool != nil {
		dbPool.Close()
	}
}

func setupDB() {
	executeSQL(createBlockInfoTableSQL())
	executeSQL(createConsensusParamsTableSQL())
	executeSQL(createMessagesTablesSQL())
}

func executeSQL(sqlStatement string) {
	if _, err := dbPool.Exec(context.Background(), sqlStatement); err != nil {
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
		height BIGINT PRIMARY KEY,
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


// CREATE TABLE IF NOT EXISTS block_txs (
// 	height BIGINT,
// 	encoded_tx TEXT,
// 	FOREIGN KEY (height) REFERENCES block_info(height)
// );
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

func createMessagesTablesSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		height BIGINT,
		type VARCHAR(255),
		sender VARCHAR(255),
		data JSONB,
		FOREIGN KEY (height) REFERENCES block_info(height),
		CONSTRAINT "messages_height_data" UNIQUE ("height", "data")
	);

	CREATE TABLE IF NOT EXISTS topics (
		id INT PRIMARY KEY,
		creator VARCHAR(255),
		metadata VARCHAR(255),
		loss_logic VARCHAR(255),
		loss_method VARCHAR(255),
		inference_logic VARCHAR(255),
		inference_method VARCHAR(255),
		epoch_last_ended VARCHAR(255),
		epoch_length VARCHAR(255),
		ground_truth_lag VARCHAR(255),
		default_arg VARCHAR(255),
		pnorm VARCHAR(255),
		alpha_regret VARCHAR(255),
		preward_reputer VARCHAR(255),
		preward_inference VARCHAR(255),
		preward_forecast VARCHAR(255),
		f_tolerance VARCHAR(255),
		message_height INT,
		message_id INT,
		FOREIGN KEY (message_height) REFERENCES block_info(height),
		FOREIGN KEY (message_id) REFERENCES messages(id)
	);

	CREATE TABLE IF NOT EXISTS addresses (
		id SERIAL PRIMARY KEY,
		pub_key VARCHAR(255) NULL DEFAULT null,
		type VARCHAR(255) NULL DEFAULT null,
		memo VARCHAR(255) NULL DEFAULT null,
		address VARCHAR(255) NULL DEFAULT null,
		CONSTRAINT "addresses_address" UNIQUE ("address"),
		CONSTRAINT "addresses_pub_key" UNIQUE ("pub_key")
	);

	CREATE TABLE IF NOT EXISTS worker_registrations (
		message_height INT,
		message_id INT,
		topic_id INT,
		sender VARCHAR(255),
		owner VARCHAR(255),
		worker_libp2pkey VARCHAR(255),
		is_reputer BOOLEAN,
		FOREIGN KEY (message_height) REFERENCES block_info(height),
		FOREIGN KEY (message_id) REFERENCES messages(id),
		FOREIGN KEY (topic_id) REFERENCES topics(id),
		FOREIGN KEY (sender) REFERENCES addresses(address),
		FOREIGN KEY (owner) REFERENCES addresses(address),
		FOREIGN KEY (worker_libp2pkey) REFERENCES addresses(pub_key)
	);

	CREATE TABLE IF NOT EXISTS transfers (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		from_address VARCHAR(255),
		topic_id INT NULL DEFAULT null,
		to_address VARCHAR(255) NULL DEFAULT null,
		amount VARCHAR(255),
		denom VARCHAR(255),
		FOREIGN KEY (message_height) REFERENCES block_info(height),
		FOREIGN KEY (message_id) REFERENCES messages(id),
		FOREIGN KEY (topic_id) REFERENCES topics(id),
		FOREIGN KEY (from_address) REFERENCES addresses(address),
		FOREIGN KEY (to_address) REFERENCES addresses(address)
	);

	CREATE TABLE IF NOT EXISTS inferences (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		nonce_block_height INT,
		topic_id INT,
		block_height INT,
		inferer VARCHAR(255),
		value VARCHAR(255),
		extra_data TEXT,
		proof TEXT,
		FOREIGN KEY (message_height) REFERENCES block_info(height),
		FOREIGN KEY (nonce_block_height) REFERENCES block_info(height),
		FOREIGN KEY (block_height) REFERENCES block_info(height),
		FOREIGN KEY (message_id) REFERENCES messages(id),
		FOREIGN KEY (topic_id) REFERENCES topics(id),
		FOREIGN KEY (inferer) REFERENCES addresses(address)
	);

	CREATE TABLE IF NOT EXISTS forcasts (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		nonce_block_height INT,
		topic_id INT,
		block_height INT,
		forcaster VARCHAR(255),
		FOREIGN KEY (message_height) REFERENCES block_info(height),
		FOREIGN KEY (nonce_block_height) REFERENCES block_info(height),
		FOREIGN KEY (block_height) REFERENCES block_info(height),
		FOREIGN KEY (message_id) REFERENCES messages(id),
		FOREIGN KEY (topic_id) REFERENCES topics(id),
		FOREIGN KEY (forcaster) REFERENCES addresses(address)
	);

	CREATE TABLE IF NOT EXISTS forcast_values (
		id BIGINT PRIMARY KEY,
		forcast_id INT,
		value VARCHAR(255),
		inferer VARCHAR(255),
		FOREIGN KEY (inferer) REFERENCES addresses(address),
		FOREIGN KEY (forcast_id) REFERENCES forcasts(id)
	);
	`


	// CREATE TABLE IF NOT EXISTS signer_infos (
	// 	id SERIAL PRIMARY KEY,
	// 	auth_info_id INT,
	// 	public_key_id INT,
	// 	sequence VARCHAR(255),
	// 	FOREIGN KEY (auth_info_id) REFERENCES auth_info(id)
	// );

	// CREATE TABLE IF NOT EXISTS public_keys (
	// 	id SERIAL PRIMARY KEY,
	// 	type VARCHAR(255),
	// 	key TEXT
	// );

	// CREATE TABLE IF NOT EXISTS auth_info (
	// 	id SERIAL PRIMARY KEY,
	// 	gas_limit VARCHAR(255),
	// 	payer VARCHAR(255),
	// 	granter VARCHAR(255)
	// 	-- Note: Tip and Amount handling depends on their structure and is omitted here
	// );
	// CREATE TABLE IF NOT EXISTS transactions (
	// 	id SERIAL PRIMARY KEY,
	// 	body_id INT,
	// 	auth_info_id INT,
	// 	signature TEXT,
	// 	FOREIGN KEY (body_id) REFERENCES messages(id),
	// 	FOREIGN KEY (auth_info_id) REFERENCES auth_info(id)
	// );


}

func insertBlockInfo(blockInfo DBBlockInfo) error {
	_, err := dbPool.Exec(context.Background(), `
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

func insertMessage(height uint64, mtype string, sender string, data string) (uint64, error) {
	// Write Topic to the database
	var id uint64
	err := dbPool.QueryRow(context.Background(), `
		INSERT INTO messages (
			height,
			type,
			sender,
			data
		) VALUES ($1, $2, $3, $4) RETURNING id`,
		height,
		mtype,
		sender,
		data,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func insertConsensusParams(params DBConsensusParams) error {
	_, err := dbPool.Exec(context.Background(), `
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

func isUniqueViolation(err error) bool {
	// This function depends on your database driver
	// For example, with PostgreSQL using pq driver:
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		println("pgErr.Code: ", pgErr.Code)
		return pgErr.Code == "23505" // 23505 is the code for unique violation in PostgreSQL
	}
	return false
}
