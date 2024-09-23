CREATE TABLE IF NOT EXISTS ema_scores (
    id SERIAL PRIMARY KEY,
    height_tx BIGINT,
    height BIGINT,
    topic_id INT,
    type VARCHAR(255),
    address VARCHAR(255),
    score NUMERIC(72,18),
    is_active BOOLEAN,
    CONSTRAINT unique_ema_score_entry UNIQUE (topic_id, type, address)
);
CREATE TABLE IF NOT EXISTS last_commit_values (
    id SERIAL PRIMARY KEY,
    height_tx BIGINT,
    height BIGINT,
    topic_id INT,
    is_worker BOOLEAN,
    CONSTRAINT unique_actor_last_commit_entry UNIQUE (topic_id, is_worker)
);
CREATE TABLE IF NOT EXISTS tokenomics (
    id SERIAL PRIMARY KEY,
    height_tx BIGINT,
    staked_amount NUMERIC(72,18),
    circulating_supply NUMERIC(72,18),
    emissions_amount NUMERIC(72,18),
    ecosystem_mint_amount NUMERIC(72,18)
);
CREATE TABLE IF NOT EXISTS topic_rewards (
    id SERIAL PRIMARY KEY,
    height_tx BIGINT,
    topic_id INT,
    reward VARCHAR(255),
    CONSTRAINT unique_topic_rewards_entry UNIQUE (topic_id, height_tx)
);
CREATE TABLE IF NOT EXISTS topic_forecasting_scores (
    id SERIAL PRIMARY KEY,
    height_tx BIGINT,
    topic_id INT,
    score VARCHAR(255),
    CONSTRAINT unique_topic_forecasting_scores_entry UNIQUE (topic_id, height_tx)
);
