-- Index for TB_SCORES on topic_id
CREATE INDEX idx_scores_topic_id ON scores (topic_id);

-- Index for TB_REWARDS on topic_id
CREATE INDEX idx_rewards_topic_id ON rewards (topic_id);

-- Index for TB_NETWORKLOSSES on topic_id
CREATE INDEX idx_networklosses_topic_id ON networklosses (topic_id);

-- Index for TB_EMASCORES on topic_id
CREATE INDEX idx_emascores_topic_id ON ema_scores (topic_id);

-- Index for TB_ACTOR_LAST_COMMIT on topic_id
CREATE INDEX idx_actor_last_commit_topic_id ON last_commit_values (topic_id);

-- Index for TB_TOPIC_REWARD on topic_id
CREATE INDEX idx_topic_reward_topic_id ON topic_rewards (topic_id);

-- Index for TB_TOPIC_FORECASTING_SCORES on topic_id
CREATE INDEX idx_topic_forecasting_scores_topic_id ON topic_forecasting_scores (topic_id);

-- Index for TB_WORKER_REGISTRATIONS on topic_id
CREATE INDEX idx_worker_registrations_topic_id ON worker_registrations (topic_id);

-- Index for TB_TRANSFERS on topic_id
CREATE INDEX idx_transfers_topic_id ON transfers (topic_id);

-- Index for TB_INFERENCES on topic_id
CREATE INDEX idx_inferences_topic_id ON inferences (topic_id);

-- Index for TB_FORECASTS on topic_id
CREATE INDEX idx_forecasts_topic_id ON forecasts (topic_id);

-- Index for TB_REPUTER_PAYLOAD on topic_id
CREATE INDEX idx_reputer_payload_topic_id ON reputer_payload (topic_id);

-- Index for TB_REPUTER_BUNDLES on topic_id
CREATE INDEX idx_reputer_bundles_topic_id ON reputer_bundles (topic_id);
