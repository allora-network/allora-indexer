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