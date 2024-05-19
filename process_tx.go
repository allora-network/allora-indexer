package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/allora-network/allora-cosmos-pump/types"
	"github.com/rs/zerolog/log"
)

// func processTxs(config ClientConfig) {

// 	// Use sql.NullInt64 which can handle NULL values
// 	var encTx string
// 	var height uint64
// 	rows, err := dbPool.Query(context.Background(), "SELECT height, encoded_tx FROM block_txs LIMIT 10")
// 	if err != nil {
// 		fmt.Errorf("failed to query SELECT encoded_tx FROM block_txs LIMIT 10: %v", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		err := rows.Scan(&height, &encTx)
// 		if err != nil {
// 			log.Fatal().Err(err)
// 		}
// 		// decodedTxData, err := decodeTx(config, encTx)
// 		// if err != nil {
// 		// 	log.Error().Err(err).Msg("Failed to process transaction")
// 		// }
// 		processTx(config, height, encTx)
// 	}

// }

func processTx(wg *sync.WaitGroup, height uint64, txData string) {
	// wgTxs.Add(1)
	defer wg.Done()
	// Decode the transaction using the decodeTx function

	txMessage, err := ExecuteCommandByKey[types.Tx](config, "decodeTx", txData)
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
			return
		}

		var messageId uint64
		messageId, err = insertMessage(height, mtype, creator, string(mjson))
		if err != nil {
			log.Error().Err(err).Msg("Failed to insertMessage")
		}

		switch mtype {
		case "/emissions.v1.MsgCreateNewTopic":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgCreateNewTopic...")
			// Add your processing logic here
			var topicPayload types.Topic
			json.Unmarshal(mjson, &topicPayload)
			insertTopic(height, messageId, topicPayload)

		case "/emissions.v1.MsgFundTopic", "/emissions.v1.MsgAddStake":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgFundTopic...")
			// Add your processing logic here
			var msgFundTopic types.MsgFundTopic
			json.Unmarshal(mjson, &msgFundTopic)
			insertMsgFundTopic(height, messageId, msgFundTopic)

		case "/cosmos.bank.v1beta1.MsgSend":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgSend...")
			// Add your processing logic here
			var msgSend types.MsgSend
			json.Unmarshal(mjson, &msgSend)
			insertMsgSend(height, messageId, msgSend)

		case "/emissions.v1.MsgInsertBulkWorkerPayload":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgInsertBulkWorkerPayload...")
			var workerPayload types.InsertBulkWorkerPayload
			json.Unmarshal(mjson, &workerPayload)
			insertInferenceForcasts(height, messageId, workerPayload)

		case "/emissions.v1.MsgRegister":
			// Process MsgProcessInferences
			log.Info().Msg("Processing MsgRegister...")
			var msgRegister types.MsgRegister
			json.Unmarshal(mjson, &msgRegister)
			insertMsgRegister(height, messageId, msgRegister)

	//! WIP below                                                  

		default:
			log.Info().Str("type", mtype).Msg("Unknown message type")
		}
	}
}

func insertMsgRegister(height uint64, messageId uint64, msg types.MsgRegister) error {
	err := insertAddress("allora", sql.NullString{msg.Sender, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgRegister insertAddress")
		return err
	}
	err = insertAddress("allora", sql.NullString{msg.Owner, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgRegister insertAddress")
		return err
	}
	err = insertAddress("libp2p", sql.NullString{msg.MultiAddress, true}, sql.NullString{msg.LibP2pKey, true}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgRegister insertAddress")
		return err
	}

	_, err = dbPool.Exec(context.Background(), `
		INSERT INTO worker_registrations (
			message_height,
			message_id,
			sender,
			topic_id,
			owner,
			worker_libp2pkey,
			is_reputer
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		height, messageId, msg.Sender, msg.TopicID, msg.Owner, msg.LibP2pKey, msg.IsReputer,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgRegister")
		return err
	}
	return nil
}

func insertAddress(t string, address sql.NullString, pub_key sql.NullString, memo string) error {
	_, err := dbPool.Exec(context.Background(), `
		INSERT INTO addresses (
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
	_, err := dbPool.Exec(context.Background(), `
		INSERT INTO transfers (
			message_height,
			message_id,
			from_address,
			topic_id,
			amount,
			denom
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		height, messageId, msg.Sender, msg.TopicID, msg.Amount, "uallo",
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgFundTopic")
		return err
	}
	return nil
}
func insertMsgSend(height uint64, messageId uint64, msg types.MsgSend) error {

	err :=insertAddress("allora", sql.NullString{msg.FromAddress, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgSend insertAddress")
		return err
	}
	err =insertAddress("allora", sql.NullString{msg.ToAddress, true}, sql.NullString{"", false}, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert insertMsgSend insertAddress")
		return err
	}
	_, err = dbPool.Exec(context.Background(), `
		INSERT INTO transfers (
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


func insertInferenceForcasts(blockHeight uint64, messageId uint64, inf types.InsertBulkWorkerPayload) error {

	for _, bundle := range inf.WorkerDataBundles {
		// Insert Forcasts
		if _, err := strconv.ParseFloat(bundle.InferenceForecastsBundle.Inference.Value, 64); err == nil {
			_, err := dbPool.Exec(context.Background(), `
				INSERT INTO inferences (
					message_height,
					message_id,
					nonce_block_height, 
					topic_id, 
					block_height,
					inferer, 
					value,
					extra_data,
					proof
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				blockHeight, messageId, inf.Nonce.BlockHeight, inf.TopicID,
				bundle.InferenceForecastsBundle.Inference.BlockHeight, bundle.InferenceForecastsBundle.Inference.Inferer,
				bundle.InferenceForecastsBundle.Inference.Value, bundle.InferenceForecastsBundle.Inference.ExtraData,
				bundle.InferenceForecastsBundle.Inference.Proof,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert inferences")
				return err
			}
		}	
		// Insert Forcasts
		if len(bundle.InferenceForecastsBundle.Forecast.ForecastElements) > 0 {
			var forcastId uint64
			err := dbPool.QueryRow(context.Background(), `
				INSERT INTO forcasts (
					message_height,
					message_id,
					nonce_block_height,
					topic_id,
					block_height,
					extra_data,
					forecaster
				) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				blockHeight, messageId, inf.Nonce.BlockHeight, inf.TopicID,
				bundle.InferenceForecastsBundle.Forecast.BlockHeight, bundle.InferenceForecastsBundle.Forecast.ExtraData,
				bundle.InferenceForecastsBundle.Forecast.Forecaster,
			).Scan(&forcastId)
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert forcasts")
				return err
			}
			for _, forecast := range bundle.InferenceForecastsBundle.Forecast.ForecastElements {
				_, err := dbPool.Exec(context.Background(), `
					INSERT INTO forcast_values (
						forecast_id,
						inferer,
						value
					) VALUES ($1, $2, $3)`,
					forcastId, forecast.Inferer, forecast.Value,
				)
				if err != nil {
					log.Error().Err(err).Msg("Failed to insert forcast_values")
					return err
				}
			}
		}

		if bundle.InferenceForecastsBundle.Inference.TopicID != inf.TopicID {
			log.Error().Msgf("Message TopicID not equal inference TopicID!!!!")
		}


	}

	return nil
}
