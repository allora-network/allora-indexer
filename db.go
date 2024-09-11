package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/allora-network/allora-indexer/types"
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

const (
	TB_EVENTS                    = "events"
	TB_MESSAGES                  = "messages"
	TB_BLOCK_INFO                = "block_info"
	TB_CONSENSUS_PARAMS          = "consensus_params"
	TB_TOPICS                    = "topics"
	TB_ADDRESSES                 = "addresses"
	TB_WORKER_REGISTRATIONS      = "worker_registrations"
	TB_TRANSFERS                 = "transfers"
	TB_INFERENCES                = "inferences"
	TB_FORECASTS                 = "forecasts"
	TB_FORECAST_VALUES           = "forecast_values"
	TB_REPUTER_PAYLOAD           = "reputer_payload"
	TB_REPUTER_BUNDLES           = "reputer_bundles"
	TB_BUNDLE_VALUES             = "bundle_values"
	TB_REWARDS                   = "rewards"
	TB_SCORES                    = "scores"
	TB_NETWORKLOSSES             = "networklosses"
	TB_NETWORKLOSS_BUNDLE_VALUES = "networkloss_bundle_values"
)

var dbPool *pgxpool.Pool //*pgx.Conn

func verifyUri(originalPath string) string {
	var res = ""
	// Split the URL into components at the '@' symbol
	parts := strings.Split(originalPath, "@")
	if len(parts) != 2 {
		fmt.Println("Invalid URL format")
		return res
	}

	// Extract the userinfo (username:password) part by splitting the protocol part
	protocolAndUserInfo := strings.Split(parts[0], "//")
	if len(protocolAndUserInfo) != 2 {
		fmt.Println("Invalid userinfo format")
		return res
	}

	protocol := protocolAndUserInfo[0]
	userInfo := protocolAndUserInfo[1]

	// Now, split the userInfo into username and password using the first ':' as a delimiter
	credParts := strings.SplitN(userInfo, ":", 2)
	if len(credParts) != 2 {
		fmt.Println("Invalid credentials format")
		return res
	}

	username := credParts[0]
	password := credParts[1]

	// Encode the password
	encodedPassword := url.QueryEscape(password)

	// Reconstruct the new userInfo part
	newUserInfo := fmt.Sprintf("%s:%s", username, encodedPassword)

	// Reconstruct the final URL
	res = fmt.Sprintf("%s//%s@%s", protocol, newUserInfo, parts[1])
	return res
}

func initDB(dataSourceName string) error {
	var err error
	// dbPool, err = pgx.Connect(context.Background(), dataSourceName)

	dbConfig, err := pgxpool.ParseConfig(verifyUri(dataSourceName))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a config, error: ")
		return err
	}
	dbPool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return err
	}
	return nil
}

func closeDB() {
	if dbPool != nil {
		dbPool.Close()
	}
}

func setupDB() error {
	err := executeSQL(createBlockInfoTableSQL())
	if err != nil {
		return err
	}
	err = executeSQL(createConsensusParamsTableSQL())
	if err != nil {
		return err
	}
	err = executeSQL(createMessagesTablesSQL())
	if err != nil {
		return err
	}
	err = executeSQL(createEventsTablesSQL())
	if err != nil {
		return err
	}
	err = addUniqueConstraints()
	if err != nil {
		return err
	}
	return nil
}

func executeSQL(sqlStatement string) error {
	if _, err := dbPool.Exec(context.Background(), sqlStatement); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute SQL statement: %v\n", err)
		return err
	}
	return nil
}

func createBlockInfoTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS ` + TB_BLOCK_INFO + ` (
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

func createConsensusParamsTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS ` + TB_CONSENSUS_PARAMS + ` (
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
	CREATE TABLE IF NOT EXISTS ` + TB_MESSAGES + ` (
		id SERIAL PRIMARY KEY,
		height BIGINT,
		type VARCHAR(255),
		sender VARCHAR(255),
		data JSONB,
		hash NUMERIC
	);

	CREATE TABLE IF NOT EXISTS ` + TB_TOPICS + ` (
		id INT PRIMARY KEY,
		creator VARCHAR(255),
		metadata VARCHAR(255),
		loss_logic VARCHAR(255),
		loss_method VARCHAR(255),
		inference_logic VARCHAR(255),
		inference_method VARCHAR(255),
		epoch_length VARCHAR(255),
		ground_truth_lag VARCHAR(255),
		default_arg VARCHAR(255),
		pnorm VARCHAR(255),
		alpha_regret VARCHAR(255),
		preward_reputer VARCHAR(255),
		preward_inference VARCHAR(255),
		preward_forecast VARCHAR(255),
		f_tolerance VARCHAR(255),
		allow_negative BOOLEAN,
		message_height INT,
		message_id INT
	);

	CREATE TABLE IF NOT EXISTS ` + TB_ADDRESSES + ` (
		id SERIAL PRIMARY KEY,
		pub_key VARCHAR(255) NULL DEFAULT null,
		type VARCHAR(255) NULL DEFAULT null,
		memo VARCHAR(255) NULL DEFAULT null,
		address VARCHAR(255) NULL DEFAULT null
	);

	CREATE TABLE IF NOT EXISTS ` + TB_WORKER_REGISTRATIONS + ` (
		message_height INT,
		message_id INT,
		topic_id INT,
		sender VARCHAR(255),
		owner VARCHAR(255),
		worker_libp2pkey VARCHAR(255),
		is_reputer BOOLEAN
	);

	CREATE TABLE IF NOT EXISTS ` + TB_TRANSFERS + ` (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		from_address VARCHAR(255),
		topic_id INT NULL DEFAULT null,
		to_address VARCHAR(255) NULL DEFAULT null,
		amount VARCHAR(255),
		denom VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS ` + TB_INFERENCES + ` (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		nonce_block_height INT,
		topic_id INT,
		block_height INT,
		inferer VARCHAR(255),
		value TEXT,
		extra_data TEXT,
		proof TEXT
	);

	CREATE TABLE IF NOT EXISTS ` + TB_FORECASTS + ` (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		nonce_block_height INT,
		topic_id INT,
		block_height INT,
		forecaster VARCHAR(255),
		extra_data VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS ` + TB_FORECAST_VALUES + ` (
		forecast_id INT,
		value VARCHAR(255),
		inferer VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS ` + TB_REPUTER_PAYLOAD + ` (
		id SERIAL PRIMARY KEY,
		message_height INT,
		message_id INT,
		sender VARCHAR(255),
		worker_nonce_block_height INT,
		reputer_nonce_block_height INT,
		topic_id INT
	);

	CREATE TABLE IF NOT EXISTS ` + TB_REPUTER_BUNDLES + ` (
		id SERIAL PRIMARY KEY,
		reputer_payload_id INT,
		pubkey VARCHAR(255),
		signature VARCHAR(255),
		reputer  VARCHAR(255),
		topic_id    INT,
		extra_data  VARCHAR(255),
		naive_value  VARCHAR(255),
		combined_value    VARCHAR(255),
		reputer_request_worker_nonce  INT,
		reputer_request_reputer_nonce  INT
	);

	DO $$ BEGIN
		CREATE TYPE reputerValueType AS ENUM(
			'InfererValues',
			'ForecasterValues',
			'OneOutInfererValues',
			'OneInForecasterValues',
			'OneOutForecasterValues'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	CREATE TABLE IF NOT EXISTS ` + TB_BUNDLE_VALUES + ` (
		bundle_id INT,
		reputer_value_type reputerValueType,
		value VARCHAR(255),
		worker VARCHAR(255)
	);`

	// FOREIGN KEY (block_height) REFERENCES block_info(height),
	// FOREIGN KEY (block_height) REFERENCES block_info(height),

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

func createEventsTablesSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS ` + TB_EVENTS + ` (
		id SERIAL PRIMARY KEY,
		height_tx BIGINT,
		height BIGINT,
		type VARCHAR(255),
		sender VARCHAR(255),
		data JSONB,
		hash NUMERIC
	);


	CREATE TABLE IF NOT EXISTS ` + TB_SCORES + ` (
		id SERIAL PRIMARY KEY,
		height_tx BIGINT,
		height BIGINT,
		topic_id INT,
		type VARCHAR(255),
		address VARCHAR(255),
		value NUMERIC(72,18),
		CONSTRAINT unique_score_entry UNIQUE (height, topic_id, type, address)
	);

	CREATE TABLE IF NOT EXISTS ` + TB_REWARDS + ` (
		id SERIAL PRIMARY KEY,
		height_tx BIGINT,
		height BIGINT,
		topic_id INT,
		type VARCHAR(255),
		address VARCHAR(255),
		value NUMERIC(72,18),
		CONSTRAINT unique_reward_entry UNIQUE (height, topic_id, type, address)
	);

	CREATE TABLE IF NOT EXISTS ` + TB_NETWORKLOSSES + ` (
		id SERIAL PRIMARY KEY,
		height_tx BIGINT,
		height BIGINT,
		topic_id INT,
		naive_value VARCHAR(255),
		combined_value VARCHAR(255),
		CONSTRAINT unique_networkloss_entry UNIQUE (height_tx, height, topic_id)
	);

	DO $$ BEGIN
		CREATE TYPE networklossBundleValueType AS ENUM(
			'InfererValues',
			'ForecasterValues',
			'OneOutInfererValues',
			'OneInForecasterValues',
			'OneOutForecasterValues'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;
	
	CREATE TABLE IF NOT EXISTS ` + TB_NETWORKLOSS_BUNDLE_VALUES + ` (
		bundle_id INT,
		reputer_value_type networklossBundleValueType,
		value VARCHAR(255),
		worker VARCHAR(255)
	);
	`
}

func insertBlockInfo(blockInfo DBBlockInfo) error {
	_, err := dbPool.Exec(context.Background(), `
		INSERT INTO `+TB_BLOCK_INFO+` (
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
	var dataHash = hash(data)
	log.Info().Msgf("Inserting message, hash: %d", height)
	err := dbPool.QueryRow(context.Background(), `
		INSERT INTO `+TB_MESSAGES+` (
			height,
			type,
			sender,
			data,
			hash
		) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		height,
		mtype,
		sender,
		data,
		dataHash,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func insertConsensusParams(params DBConsensusParams) error {
	_, err := dbPool.Exec(context.Background(), `
        INSERT INTO `+TB_CONSENSUS_PARAMS+` (
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
		return pgErr.Code == "23505" // 23505 is the code for unique violation in PostgreSQL
	}
	return false
}

// Events
type EventRecord struct {
	Height uint64
	Type   string
	Sender string
	Data   json.RawMessage
}

func isEventType(eventType, prefix, suffix string) bool {
	return strings.HasPrefix(eventType, prefix) && strings.HasSuffix(eventType, suffix)
}

// isScoreEvent checks if the event is a score event based on its type.
func isScoreEvent(event EventRecord) bool {
	return isEventType(event.Type, "emissions.v", "EventScoresSet")
}

// isRewardEvent checks if the event is a reward event based on its type.
func isRewardEvent(event EventRecord) bool {
	return isEventType(event.Type, "emissions.v", "EventRewardsSettled")
}

// isNetworkLossEvent checks if the event is a network loss event based on its type.
func isNetworkLossEvent(event EventRecord) bool {
	return isEventType(event.Type, "emissions.v", "EventNetworkLossSet")
}

func insertEvents(events []EventRecord) error {
	var scoreEvents []EventRecord
	var rewardEvents []EventRecord
	var networkLossEvents []EventRecord
	// For inserting events in batch:
	var insertStatements []string
	var values []interface{}
	placeholderCounter := 1 // Placeholder index starts at 1 in PostgreSQL

	for _, event := range events {
		// Determine the type of event and accumulate accordingly
		if isScoreEvent(event) { // Function to check if it's a score event
			scoreEvents = append(scoreEvents, event)
		} else if isRewardEvent(event) { // Function to check if it's a reward event
			rewardEvents = append(rewardEvents, event)
		} else if isNetworkLossEvent(event) { // Function to check if it's a network loss event
			networkLossEvents = append(networkLossEvents, event)
		} else {
			log.Info().Msg("Unrecognized event, ignoring")
			continue
		}

		// Prepare data for batch insert into TB_EVENTS
		dataHash := hash(string(event.Data))
		newStmt := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", placeholderCounter, placeholderCounter+1, placeholderCounter+2, placeholderCounter+3, placeholderCounter+4)
		insertStatements = append(insertStatements, newStmt)
		values = append(values, event.Height, event.Type, event.Sender, string(event.Data), dataHash)
		placeholderCounter += 5 // Increase counter for next row
	}

	// Batch insert into TB_EVENTS
	if len(insertStatements) > 0 {
		sqlStatement := fmt.Sprintf(`
				INSERT INTO %s (height, type, sender, data, hash) 
				VALUES %s
				ON CONFLICT (height, hash, type) DO NOTHING`, TB_EVENTS, strings.Join(insertStatements, ","))

		// Log the SQL statement and values for debugging
		log.Debug().Str("SQL Statement", sqlStatement).Interface("Values", values).Msg("Executing batch insert for events")

		_, err := dbPool.Exec(context.Background(), sqlStatement, values...)
		if err != nil {
			return fmt.Errorf("event insert failed: %v", err)
		}
	} else {
		log.Info().Msg("No events data to insert")
	}

	// Insert scores if any
	if len(scoreEvents) > 0 {
		err := insertScore(scoreEvents)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert scores")
		}
	}

	// Insert rewards if any
	if len(rewardEvents) > 0 {
		err := insertReward(rewardEvents)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert rewards")
		}
	}

	// Insert network losses if any
	if len(networkLossEvents) > 0 {
		err := insertNetworkLoss(networkLossEvents)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert network losses")
		}
	}

	return nil
}

func insertScore(events []EventRecord) error {
	log.Info().Msg("Inserting scores in batch")
	var insertStatements []string
	var values []interface{}

	placeholderCounter := 1 // Placeholder index starts at 1 in PostgreSQL

	for _, event := range events {
		log.Trace().Interface("Event score", event).Msg("Processing event score")
		var attributes []Attribute
		err := json.Unmarshal(event.Data, &attributes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		var topicID int
		var actorType string
		var addresses []string
		var scores []big.Float
		var blockHeight int

		for _, attr := range attributes {
			switch attr.Key {
			case "topic_id":
				cleanedValue := strings.Trim(attr.Value, "\"")
				topicID, err = strconv.Atoi(cleanedValue)
				if err != nil {
					return fmt.Errorf("failed to convert topic_id to int: %w", err)
				}
			case "actor_type":
				actorType = strings.Trim(attr.Value, "\"")
			case "block_height":
				cleanedValue := strings.Trim(attr.Value, "\"")
				blockHeight, err = strconv.Atoi(cleanedValue)
				if err != nil {
					return fmt.Errorf("failed to convert block_height to int: %w", err)
				}
			case "addresses":
				err = json.Unmarshal([]byte(attr.Value), &addresses)
				if err != nil {
					return fmt.Errorf("failed to unmarshal addresses: %w", err)
				}
			case "scores":
				var rawScores []string
				err = json.Unmarshal([]byte(attr.Value), &rawScores)
				if err != nil {
					return fmt.Errorf("failed to unmarshal scores: %w", err)
				}

				for _, rawScore := range rawScores {
					rawScoreClean := strings.Trim(rawScore, "\"")
					if isInvalidNumericValue(rawScoreClean) {
						log.Error().Str("rawScore", rawScore).Msg("Failed to convert score to big.Float")
						return fmt.Errorf("Invalid Score: %s", rawScoreClean)
					} else {
						score := new(big.Float)
						score, ok := score.SetString(rawScoreClean)
						if !ok {
							log.Error().Str("rawScore", rawScore).Msg("Failed to convert score to big.Float")
							return fmt.Errorf("Invalid Score: %s", rawScoreClean)
						}
						scores = append(scores, *score)
					}
				}
			}
		}

		if len(addresses) != len(scores) {
			return fmt.Errorf("mismatch in length of addresses and scores")
		}

		for i := range addresses {
			// Generate the placeholders for this row
			newStmt := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", placeholderCounter, placeholderCounter+1, placeholderCounter+2, placeholderCounter+3, placeholderCounter+4, placeholderCounter+5)
			insertStatements = append(insertStatements, newStmt)
			scoreValue := scores[i].Text('f', -1)
			values = append(values, event.Height, blockHeight, topicID, actorType, addresses[i], scoreValue)
			placeholderCounter += 6 // Increase counter for next row
		}
	}

	if len(insertStatements) > 0 {
		sqlStatement := fmt.Sprintf(`
			INSERT INTO %s (height_tx, height, topic_id, type, address, value) 
			VALUES %s`, TB_SCORES, strings.Join(insertStatements, ","))
		log.Trace().Str("Event - Score SQL Statement", sqlStatement).Interface("Values", values).Msg("Executing batch insert for scores")
		_, err := dbPool.Exec(context.Background(), sqlStatement, values...)
		if err != nil {
			return fmt.Errorf("score insert failed: %v", err)
		}
	} else {
		log.Info().Msg("No scores data to insert")
	}

	return nil
}

func insertReward(events []EventRecord) error {
	log.Info().Msg("Inserting rewards in batch")
	var insertStatements []string
	var values []interface{}
	placeholderCounter := 1 // Placeholder index starts at 1 in PostgreSQL

	for _, event := range events {
		log.Trace().Interface("Event reward", event).Msg("Processing event reward")
		var attributes []Attribute
		err := json.Unmarshal(event.Data, &attributes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		var topicID int
		var rewardType string
		var addresses []string
		var rewards []big.Float
		var blockHeight int

		for _, attr := range attributes {
			switch attr.Key {
			case "topic_id":
				cleanedValue := strings.Trim(attr.Value, "\"")
				topicID, err = strconv.Atoi(cleanedValue)
				if err != nil {
					return fmt.Errorf("failed to convert topic_id to int: %w", err)
				}
			case "actor_type":
				rewardType = strings.Trim(attr.Value, "\"")
			case "block_height":
				cleanedValue := strings.Trim(attr.Value, "\"")
				blockHeight, err = strconv.Atoi(cleanedValue)
				if err != nil {
					return fmt.Errorf("failed to convert block_height to int: %w", err)
				}
			case "addresses":
				err = json.Unmarshal([]byte(attr.Value), &addresses)
				if err != nil {
					return fmt.Errorf("failed to unmarshal addresses: %w", err)
				}
			case "rewards":
				err = json.Unmarshal([]byte(attr.Value), &rewards)
				if err != nil {
					return fmt.Errorf("failed to unmarshal rewards: %w", err)
				}
			}
		}

		if len(addresses) != len(rewards) {
			return fmt.Errorf("mismatch in length of addresses and rewards")
		}

		for i := range addresses {
			// Generate the placeholders for this row
			newStmt := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", placeholderCounter, placeholderCounter+1, placeholderCounter+2, placeholderCounter+3, placeholderCounter+4, placeholderCounter+5)
			insertStatements = append(insertStatements, newStmt)
			values = append(values, event.Height, blockHeight, topicID, rewardType, addresses[i], rewards[i].Text('f', -1))
			placeholderCounter += 6 // Increase counter for next row
		}
	}

	if len(insertStatements) > 0 {
		sqlStatement := fmt.Sprintf(`
			INSERT INTO %s (height_tx, height, topic_id, type, address, value) 
			VALUES %s`, TB_REWARDS, strings.Join(insertStatements, ","))

		log.Trace().Str("Event - Reward SQL Statement", sqlStatement).Interface("Values", values).Msg("Executing batch insert for scores")

		_, err := dbPool.Exec(context.Background(), sqlStatement, values...)
		if err != nil {
			return fmt.Errorf("rewards insert failed: %v", err)
		}
	} else {
		log.Info().Msg("No rewards data to insert")
	}

	return nil
}

// func insertNetworkLoss(event EventRecord) error {
func insertNetworkLoss(events []EventRecord) error {
	for _, event := range events {
		log.Debug().Interface("Event network loss", event).Msg("inserting event network loss ")
		var attributes []Attribute
		err := json.Unmarshal(event.Data, &attributes)
		if err != nil {
			return err
		}

		var topicID int
		var block_height int
		var valueBundle types.MsgValueBundle

		for _, attr := range attributes {
			cleanedValue := strings.Trim(attr.Value, "\"")
			switch attr.Key {
			case "topic_id":
				topicID, err = strconv.Atoi(cleanedValue)
				if err != nil {
					return err
				}
			case "block_height":
				block_height, err = strconv.Atoi(cleanedValue)
				if err != nil {
					return err
				}
			case "value_bundle":
				err = json.Unmarshal([]byte(cleanedValue), &valueBundle)
				if err != nil {
					return err
				}
			}
		}

		var bundleId uint64
		err = dbPool.QueryRow(context.Background(), `
				INSERT INTO `+TB_NETWORKLOSSES+` (height_tx, height, topic_id, naive_value, combined_value) VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (height_tx, height, topic_id) DO NOTHING returning id`,
			event.Height, block_height, topicID, valueBundle.NaiveValue, valueBundle.CombinedValue).Scan(&bundleId)

		if err != nil {
			return fmt.Errorf("network loss event insert failed: %v", err)
		}

		log.Debug().Msgf("Inserting NetworkLoss bundle: %d, %v", bundleId, valueBundle)
		insertValueBundle(bundleId, valueBundle, TB_NETWORKLOSS_BUNDLE_VALUES)
	}
	return nil
}

func isDataEmpty(table string) (bool, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := dbPool.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to check if table %s is empty", table)
		return false, err
	}
	return count == 0, nil
}

func tableExists(tableName string) (bool, error) {
	var exists bool
	err := dbPool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = $1
		)`, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func insertValueBundle(
	bundleId uint64,
	valueBundle types.MsgValueBundle,
	tableName string,
) error {

	//Insert InfererValues
	for _, val := range valueBundle.InfererValues {
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+tableName+` (
					bundle_id,
					reputer_value_type,
					worker,
					value
				) VALUES ($1, $2, $3, $4)`,
			bundleId, "InfererValues", val.Worker, val.Value,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert InfererValues bundle_values")
			return err
		}
	}
	//Insert ForecasterValues
	for _, val := range valueBundle.ForecasterValues {
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+tableName+` (
					bundle_id,
					reputer_value_type,
					worker,
					value
				) VALUES ($1, $2, $3, $4)`,
			bundleId, "ForecasterValues", val.Worker, val.Value,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert ForecasterValues bundle_values")
			return err
		}
	}
	// Insert OneOutInfererValues
	for _, val := range valueBundle.OneOutInfererValues {
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+tableName+` (
					bundle_id,
					reputer_value_type,
					worker,
					value
				) VALUES ($1, $2, $3, $4)`,
			bundleId, "OneOutInfererValues", val.Worker, val.Value,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert OneOutInfererValues bundle_values")
			return err
		}
	}
	// Insert OneInForecasterValues
	for _, val := range valueBundle.OneInForecasterValues {
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+tableName+` (
					bundle_id,
					reputer_value_type,
					worker,
					value
				) VALUES ($1, $2, $3, $4)`,
			bundleId, "OneInForecasterValues", val.Worker, val.Value,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert OneInForecasterValues bundle_values")
			return err
		}
	}
	// Insert OneOutForecasterValues
	for _, val := range valueBundle.OneOutForecasterValues {
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+tableName+` (
					bundle_id,
					reputer_value_type,
					worker,
					value
				) VALUES ($1, $2, $3, $4)`,
			bundleId, "OneOutForecasterValues", val.Worker, val.Value,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert OneOutForecasterValues bundle_values")
			return err
		}
	}
	// Insert OneOutInfererForecasterValues
	for _, val := range valueBundle.OneOutInfererForecasterValues {
		oneOutInfererStrValues := ""
		if len(val.OneOutInfererValues) != 0 {
			mjson, err := json.Marshal(val.OneOutInfererValues)
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert OneOutInfererForecasterValues bundle_values")
				return err
			}
			oneOutInfererStrValues = string(mjson)
		}
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+tableName+` (
					bundle_id,
					reputer_value_type,
					worker,
					value
				) VALUES ($1, $2, $3, $4)`,
			bundleId, "OneOutInfererForecasterValues", val.Forecaster, oneOutInfererStrValues,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert OneOutInfererForecasterValues bundle_values")
			return err
		}
	}
	return nil
}

func addUniqueConstraints() error {
	_, err := dbPool.Exec(context.Background(), `
				ALTER TABLE `+TB_MESSAGES+` drop CONSTRAINT IF EXISTS messages_height_data`,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove constraint unique from message")
	}

	_, err = dbPool.Exec(context.Background(), `
				ALTER TABLE `+TB_MESSAGES+` ADD CONSTRAINT messages_height_data UNIQUE (height, hash)`,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add constraint unique to message")
		return err
	}

	_, err = dbPool.Exec(context.Background(), `
				ALTER TABLE `+TB_EVENTS+` drop CONSTRAINT IF EXISTS events_height_data`,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove constraint unique from events")
	}

	_, err = dbPool.Exec(context.Background(), `
				ALTER TABLE `+TB_EVENTS+` ADD CONSTRAINT events_height_data UNIQUE (height, hash, type)`,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add constraint unique to events")
		return err
	}

	return nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func isInvalidNumericValue(value string) bool {
	return strings.Contains(strings.ToLower(value), "infinity") || strings.Contains(strings.ToLower(value), "nan")
}
