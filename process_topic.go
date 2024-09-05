package main

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/allora-network/allora-indexer/types"
	"github.com/rs/zerolog/log"
)

// * Process topics
type TopicOutput struct {
	Topic types.MsgCreateNewTopic `json:"topic"`
}

func insertMsgCreateNewTopic(height uint64, messageId uint64, topic types.MsgCreateNewTopic) error {

	lastTopicIdStr, err := ExecuteCommandByKey[struct {
		NextTopicID string `json:"next_topic_id"`
	}](config, "nextTopicId", "--height", strconv.FormatUint(height, 10))
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		return err
	}
	lastTopicIdStr1Back, err := ExecuteCommandByKey[struct {
		NextTopicID string `json:"next_topic_id"`
	}](config, "nextTopicId", "--height", strconv.FormatUint(height-1, 10))
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		return err
	}
	topId1Back, err := strconv.Atoi(lastTopicIdStr1Back.NextTopicID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert lastTopicIdStr1Back.NextTopicID to int")
		return err
	}
	topId, err := strconv.Atoi(lastTopicIdStr.NextTopicID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert lastTopicIdStr.NextTopicID to int")
		return err
	}

	for i := topId1Back; i < topId; i++ {
		log.Info().Msgf("%d: (%d) == (%d)", i, topId, topId1Back)

		t, err := ExecuteCommandByKey[TopicOutput](config, "topicById", strconv.Itoa(i), "--height", strconv.FormatUint(height, 10))
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute command topicById in insertTopics")
			return err
		}

		if topic.Metadata == t.Topic.Metadata {
			err := insertAddress("allora", sql.NullString{topic.Creator, true}, sql.NullString{"", false}, "")
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert insertMsgSend insertAddress")
				return err
			}
			log.Info().Msgf("Insert topic (%s, %s) into DB...", t.Topic.TopicID, t.Topic.Metadata)
			_, err = dbPool.Exec(context.Background(), `
				INSERT INTO `+TB_TOPICS+` (
					id,
					creator,
					metadata,
					loss_logic,
					loss_method,
					inference_logic,
					inference_method,
					epoch_length,
					ground_truth_lag,
					default_arg,
					pnorm,
					alpha_regret,
					preward_reputer,
					preward_inference,
					preward_forecast,
					f_tolerance,
					allow_negative,
					message_height,
					message_id
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
				t.Topic.TopicID,
				t.Topic.Creator,
				t.Topic.Metadata,
				t.Topic.LossLogic,
				t.Topic.LossMethod,
				t.Topic.InferenceLogic,
				t.Topic.InferenceMethod,
				t.Topic.EpochLength,
				t.Topic.GroundTruthLag,
				t.Topic.DefaultArg,
				t.Topic.Pnorm,
				t.Topic.AlphaRegret,
				t.Topic.PrewardReputer,
				t.Topic.PrewardInference,
				t.Topic.PrewardForecast,
				t.Topic.FTolerance,
				t.Topic.AllowNegative,
				height,
				messageId,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert topic")
				return err
			}
			return nil
		}
	}

	return nil
}
