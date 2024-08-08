package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/allora-network/allora-cosmos-pump/types"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const MAX_RETRY int = 3
const RETRY_PAUSE int = 2

func processTx(wg *sync.WaitGroup, height uint64, txData string) {
	defer wg.Done()

	// Decode the transaction using the decodeTx function
	//txMessage, err := ExecuteCommandByKey[types.Tx](config, "decodeTx", txData)
	txMessage, err := DecodeTx(config, txData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
	}

	// Process the decoded transaction message
	for _, msg := range txMessage.Body.Messages {
		mtype := msg["@type"].(string) //fmt.Sprint(msg["@type"])
		mjson, err := json.Marshal(msg)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal msg")
		}
		var creator string
		if msg["creator"] != nil {
			creator = msg["creator"].(string)
		} else if msg["sender"] != nil {
			creator = msg["sender"].(string)
		} else if msg["from_address"] != nil {
			creator = msg["from_address"].(string)
		} else {
			log.Error().Msg("Cannot define creator!!!")
		}

		var messageId uint64
		log.Info().Msgf("Inserting message, height: %d", height)
		messageId, err = insertMessage(height, mtype, creator, string(mjson))
		if err != nil {
			log.Error().Err(err).Msgf("Failed to insertMessage, height: %d", height)
		}

		switch mtype {
		case "/emissions.v1.MsgCreateNewTopic", "/emissions.v2.MsgCreateNewTopic":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgCreateNewTopic...")
			// Add your processing logic here
			var topicPayload types.MsgCreateNewTopic
			json.Unmarshal(mjson, &topicPayload)
			insertMsgCreateNewTopic(height, messageId, topicPayload)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertMsgCreateNewTopic, height: %d", height)
			}

		case "/emissions.v1.MsgFundTopic", "/emissions.v1.MsgAddStake", "/emissions.v2.MsgFundTopic", "/emissions.v2.MsgAddStake":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgFundTopic...")
			// Add your processing logic here
			var msgFundTopic types.MsgFundTopic
			json.Unmarshal(mjson, &msgFundTopic)
			insertMsgFundTopic(height, messageId, msgFundTopic)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertMsgFundTopic, height: %d", height)
			}

		case "/cosmos.bank.v1beta1.MsgSend":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgSend...")
			// Add your processing logic here
			var msgSend types.MsgSend
			json.Unmarshal(mjson, &msgSend)
			insertMsgSend(height, messageId, msgSend)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertMsgSend, height: %d", height)
			}

		case "/emissions.v1.MsgRegister", "/emissions.v2.MsgRegister":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgRegister...")
			var msgRegister types.MsgRegister
			json.Unmarshal(mjson, &msgRegister)
			insertMsgRegister(height, messageId, msgRegister)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertMsgRegister, height: %d", height)
			}

		case "/emissions.v1.MsgInsertBulkWorkerPayload":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgInsertBulkWorkerPayload...")
			var workerPayload types.MsgInsertBulkWorkerPayload
			json.Unmarshal(mjson, &workerPayload)
			insertBulkWorkerPayload(height, messageId, workerPayload)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertBulkWorkerPayload, height: %d", height)
			}
		case "/emissions.v2.MsgInsertWorkerPayload":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgInsertWorkerPayload...")
			var workerPayload types.MsgInsertWorkerPayload
			json.Unmarshal(mjson, &workerPayload)
			insertWorkerPayload(height, messageId, workerPayload)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertBulkWorkerPayload, height: %d", height)
			}

		case "/emissions.v1.MsgInsertBulkReputerPayload":
			// Process MsgInsertReputerPayload
			log.Info().Msg("Processing MsgInsertBulkReputerPayload...")
			var reputerPayload types.MsgInsertBulkReputerPayload
			json.Unmarshal(mjson, &reputerPayload)
			insertBulkReputerPayload(height, messageId, reputerPayload)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertInferenceForecasts, height: %d", height)
			}
		case "/emissions.v2.MsgInsertReputerPayload":
			// Process MsgInsertReputerPayload
			log.Info().Msg("Processing MsgInsertReputerPayload...")
			var reputerPayload types.MsgInsertReputerPayload
			json.Unmarshal(mjson, &reputerPayload)
			insertReputerPayload(height, messageId, reputerPayload)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to insertInferenceForecasts, height: %d", height)
			}

		default:
			log.Info().Str("type", mtype).Msg("Unknown message type")
		}
	}
}

func insertBulkReputerPayload(blockHeight uint64, messageId uint64, msg types.MsgInsertBulkReputerPayload) error {

	worker_nonce_block_height, err := strconv.Atoi(msg.ReputerRequestNonce.WorkerNonce.BlockHeight)
	reputer_nonce_block_height, err := strconv.Atoi(msg.ReputerRequestNonce.ReputerNonce.BlockHeight)
	var payloadId uint64
	err = dbPool.QueryRow(context.Background(), `
		INSERT INTO `+TB_REPUTER_PAYLOAD+` (
			message_height,
			message_id,
			sender,
			worker_nonce_block_height,
			reputer_nonce_block_height,
			topic_id
		) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		blockHeight, messageId, msg.Sender, worker_nonce_block_height,
		reputer_nonce_block_height, msg.TopicID,
	).Scan(&payloadId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert reputer_payload")
		return err
	}

	var bundleId uint64
	for _, bundle := range msg.ReputerValueBundles {
		log.Info().Msgf("Inserting bundle: %v", bundle)
		request_worker_nonce_block_height, err := strconv.Atoi(bundle.ValueBundle.ReputerRequestNonce.WorkerNonce.BlockHeight)
		request_reputer_nonce_block_height, err := strconv.Atoi(bundle.ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight)
		err = insertAddress("allora", sql.NullString{"", false}, sql.NullString{bundle.Pubkey, true}, "")
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert bundle.Pubkey insertAddress")
			return err
		}
		err = dbPool.QueryRow(context.Background(), `
			INSERT INTO `+TB_REPUTER_BUNDLES+` (
				reputer_payload_id,
				pubkey,
				signature,
				reputer,
				topic_id,
				extra_data,
				naive_value,
				combined_value,
				reputer_request_worker_nonce,
				reputer_request_reputer_nonce
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
			payloadId, sql.NullString{bundle.Pubkey, true}, bundle.Signature, bundle.ValueBundle.Reputer,
			bundle.ValueBundle.TopicID, bundle.ValueBundle.ExtraData, bundle.ValueBundle.NaiveValue,
			bundle.ValueBundle.CombinedValue, request_worker_nonce_block_height,
			request_reputer_nonce_block_height,
		).Scan(&bundleId)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert reputer_bundles")
			return err
		}
		err = insertValueBundle(bundleId, bundle.ValueBundle, TB_BUNDLE_VALUES)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertReputerPayload(blockHeight uint64, messageId uint64, msg types.MsgInsertReputerPayload) error {

	reputer_nonce_block_height, err := strconv.Atoi(msg.ReputerValueBundle.ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight)
	var payloadId uint64
	err = dbPool.QueryRow(context.Background(), `
		INSERT INTO `+TB_REPUTER_PAYLOAD+` (
			message_height,
			message_id,
			sender,
			reputer_nonce_block_height,
			topic_id
		) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		blockHeight, messageId, msg.Sender,
		reputer_nonce_block_height, msg.ReputerValueBundle.ValueBundle.TopicID,
	).Scan(&payloadId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert reputer_payload")
		return err
	}

	var bundleId uint64
	log.Info().Msgf("Inserting bundle: %v", msg.ReputerValueBundle)
	err = insertAddress("allora", sql.NullString{"", false}, sql.NullString{msg.ReputerValueBundle.Pubkey, true}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert bundle.Pubkey insertAddress")
		return err
	}
	err = dbPool.QueryRow(context.Background(), `
			INSERT INTO `+TB_REPUTER_BUNDLES+` (
				reputer_payload_id,
				pubkey,
				signature,
				reputer,
				topic_id,
				extra_data,
				naive_value,
				combined_value,
				reputer_request_reputer_nonce
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		payloadId, sql.NullString{msg.ReputerValueBundle.Pubkey, true}, msg.ReputerValueBundle.Signature, msg.ReputerValueBundle.ValueBundle.Reputer,
		msg.ReputerValueBundle.ValueBundle.TopicID, msg.ReputerValueBundle.ValueBundle.ExtraData, msg.ReputerValueBundle.ValueBundle.NaiveValue,
		msg.ReputerValueBundle.ValueBundle.CombinedValue,
		msg.ReputerValueBundle.ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight,
	).Scan(&bundleId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert reputer_bundles")
		return err
	}
	err = insertValueBundle(bundleId, msg.ReputerValueBundle.ValueBundle, TB_BUNDLE_VALUES)
	if err != nil {
		return err
	}

	return nil
}

func insertBulkWorkerPayload(blockHeight uint64, messageId uint64, inf types.MsgInsertBulkWorkerPayload) error {

	for _, bundle := range inf.WorkerDataBundles {

		nonce_block_height, err := strconv.Atoi(inf.Nonce.BlockHeight)
		if err != nil {
			log.Error().Err(err).Msg("Failed to convert inf.Nonce.BlockHeight to int in insertInferenceForecasts")
			return err
		}
		topic_id, err := strconv.Atoi(inf.TopicID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to convert inf.TopicID to int in insertInferenceForecasts")
			return err
		}
		block_height, err := strconv.Atoi(bundle.InferenceForecastsBundle.Inference.BlockHeight)
		if err != nil {
			log.Error().Err(err).Msg("Failed to convert bundle.InferenceForecastsBundle.Inference.BlockHeight to int in insertInferenceForecasts")
			return err
		}
		waitCreation("block_info", "height", strconv.FormatUint(blockHeight, 10))
		if err != nil {
			log.Error().Err(err).Msg("height is still not exist in block_info blockHeight. Exiting...")
			return err
		}
		err = insertWorkerDataBundle(messageId, blockHeight, block_height, topic_id, nonce_block_height, bundle)
		if err != nil {
			//log.Error().Err(err).Msg("height is still not exist in block_info blockHeight. Exiting...")
			return err
		}
	}

	return nil
}

func insertWorkerDataBundle(messageId, blockHeight uint64, block_height, topic_id, nonce_block_height int, bundle types.WorkerDataBundle) error {

	inferenceTopicId, err := strconv.Atoi(bundle.InferenceForecastsBundle.Inference.TopicID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert inference TopicId to int in insertMsgFundTopic")
		return err
	}
	// Insert inference
	log.Info().Msgf("Inserting inference nonce: %d, value: %s, topic_id: %d", nonce_block_height, bundle.InferenceForecastsBundle.Inference.Value, topic_id)
	if _, err := strconv.ParseFloat(bundle.InferenceForecastsBundle.Inference.Value, 64); err == nil {
		_, err := dbPool.Exec(context.Background(), `
				INSERT INTO `+TB_INFERENCES+` (
					message_height,
					message_id,
					nonce_block_height,
					topic_id,
					block_height,
					inferer,
					value,
					extra_data
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			blockHeight, messageId, nonce_block_height, topic_id,
			block_height, bundle.InferenceForecastsBundle.Inference.Inferer,
			bundle.InferenceForecastsBundle.Inference.Value, bundle.InferenceForecastsBundle.Inference.ExtraData,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert inferences")
			return err
		}
	} else {
		log.Error().Err(err).Msg("Failed to convert inference value")
		return err
	}
	// Insert Forecasts
	if len(bundle.InferenceForecastsBundle.Forecast.ForecastElements) > 0 {

		var forecastId uint64
		err := dbPool.QueryRow(context.Background(), `
				INSERT INTO `+TB_FORECASTS+` (
					message_height,
					message_id,
					nonce_block_height,
					topic_id,
					block_height,
					extra_data,
					forecaster
				) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			blockHeight, messageId, nonce_block_height, topic_id,
			bundle.InferenceForecastsBundle.Forecast.BlockHeight, bundle.InferenceForecastsBundle.Forecast.ExtraData,
			bundle.InferenceForecastsBundle.Forecast.Forecaster,
		).Scan(&forecastId)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert forecasts")
			return err
		}
		log.Info().Msgf("forecast_id: %d", forecastId)
		for _, forecast := range bundle.InferenceForecastsBundle.Forecast.ForecastElements {
			_, err := dbPool.Exec(context.Background(), `
					INSERT INTO `+TB_FORECAST_VALUES+` (
						forecast_id,
						inferer,
						value
					) VALUES ($1, $2, $3)`,
				forecastId, forecast.Inferer, forecast.Value,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert forecast_values")
				return err
			}
		}
	}

	if inferenceTopicId != topic_id {
		log.Error().Msgf("Message TopicID not equal inference TopicID!!!!")
	}

	return nil
}

func insertWorkerPayload(blockHeight uint64, messageId uint64, inf types.MsgInsertWorkerPayload) error {

	nonce_block_height, err := strconv.Atoi(inf.WorkerDataBundle.Nonce.BlockHeight)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert inf.Nonce.BlockHeight to int in insertInferenceForecasts")
		return err
	}
	topic_id, err := strconv.Atoi(inf.WorkerDataBundle.TopicID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert inf.TopicID to int in insertInferenceForecasts")
		return err
	}
	block_height, err := strconv.Atoi(inf.WorkerDataBundle.InferenceForecastsBundle.Inference.BlockHeight)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert bundle.InferenceForecastsBundle.Inference.BlockHeight to int in insertInferenceForecasts")
		return err
	}
	waitCreation("block_info", "height", strconv.FormatUint(blockHeight, 10))
	if err != nil {
		log.Error().Err(err).Msg("height is still not exist in block_info blockHeight. Exiting...")
		return err
	}
	err = insertWorkerDataBundle(messageId, blockHeight, block_height, topic_id, nonce_block_height, inf.WorkerDataBundle)
	if err != nil {
		return err
	}

	return nil
}

func waitCreation(table string, field string, value string) error {
	var err error
	for _ = range MAX_RETRY {
		var count int
		err = dbPool.QueryRow(context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = %s", table, field, value),
		).Scan(&count)
		if count > 0 {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	log.Error().Err(err).Msgf("Failed to get %s: %s from table: %s", field, value, table)
	return err
}

func insertMsgRegister(height uint64, messageId uint64, msg types.MsgRegister) error {
	err := insertAddress("allora", sql.NullString{msg.Sender, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgRegister insertAddress")
		return err
	}

	topId, err := strconv.Atoi(msg.TopicID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert msg.TopicID to int")
		return err
	}

	err = waitCreation("topics", "id", strconv.Itoa(topId))
	if err != nil {
		log.Error().Err(err).Msg("TopicId is still not exist in DB. Exiting...")
		return err
	}

	_, err = dbPool.Exec(context.Background(), `
		INSERT INTO `+TB_WORKER_REGISTRATIONS+` (
			message_height,
			message_id,
			sender,
			topic_id,
			owner,
			worker_libp2pkey,
			is_reputer
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		height, messageId, msg.Sender, topId, msg.Owner, msg.LibP2pKey, msg.IsReputer,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgRegister")
		return err
	}
	return nil
}

func insertAddress(t string, address sql.NullString, pub_key sql.NullString, memo string) error {
	_, err := dbPool.Exec(context.Background(), `
		INSERT INTO `+TB_ADDRESSES+` (
			pub_key,
			type,
			memo,
			address
		) VALUES ($1, $2, $3, $4)`,
		pub_key, t, memo, address,
	)
	if err != nil {
		if isUniqueViolation(err) {
			log.Info().Msgf("Address/pub_key %s/%s already exist. Skipping insert.", address.String, pub_key.String)
			return nil // or return an error if you prefer
		}
		log.Error().Err(err).Msg("Failed to insert insertAddress")
		return err
	}
	return nil
}

func insertMsgFundTopic(height uint64, messageId uint64, msg types.MsgFundTopic) error {
	topId, err := strconv.Atoi(msg.TopicID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert msg.TopicID to int in insertMsgFundTopic")
		return err
	}

	insertAddress("allora", sql.NullString{msg.Sender, true}, sql.NullString{"", false}, "")

	err = waitCreation("topics", "id", strconv.Itoa(topId))
	if err != nil {
		log.Error().Err(err).Msg("TopicId is still not exist in DB. Exiting...")
		return err
	}

	_, err = dbPool.Exec(context.Background(), `
		INSERT INTO `+TB_TRANSFERS+` (
			message_height,
			message_id,
			from_address,
			topic_id,
			amount,
			denom
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		height, messageId, msg.Sender, topId, msg.Amount, "uallo",
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgFundTopic")
		return err
	}
	return nil
}
func insertMsgSend(height uint64, messageId uint64, msg types.MsgSend) error {

	err := insertAddress("allora", sql.NullString{msg.FromAddress, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgSend insertAddress")
		return err
	}
	err = insertAddress("allora", sql.NullString{msg.ToAddress, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgSend insertAddress")
		return err
	}
	_, err = dbPool.Exec(context.Background(), `
		INSERT INTO `+TB_TRANSFERS+` (
			message_height,
			message_id,
			from_address,
			to_address,
			amount,
			denom
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		height, messageId, msg.FromAddress, msg.ToAddress, msg.Amount[0].Amount, msg.Amount[0].Denom,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgSend")
		return err
	}
	return nil
}
