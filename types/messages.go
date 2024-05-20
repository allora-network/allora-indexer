package types

type GenericMessage struct {
	Type string `json:"@type"`
}

type MessageBody struct {
	Messages []GenericMessage `json:"messages"`
}

type Transaction struct {
	Body MessageBody `json:"body"`
}

type MsgProcessInferences struct {
	Sender     string `json:"sender"`
	Inferences []struct {
		TopicID   string `json:"topic_id"`
		Worker    string `json:"worker"`
		Value     string `json:"value"`
		ExtraData any    `json:"extra_data"`
		Proof     string `json:"proof"`
	} `json:"inferences"`
}

type InsertBulkWorkerPayload struct {
	Type                  string `json:"@type"`
	Nonce                 struct {
		BlockHeight string `json:"block_height"`
	} `json:"nonce"`
	Sender               string `json:"sender"`
	TopicID              string `json:"topic_id"`
	WorkerDataBundles    []struct {
		Pubkey                      string `json:"pubkey"`
		Worker                      string `json:"worker"`
		InferenceForecastsBundle    struct {
			Forecast struct {
				TopicID     string      `json:"topic_id"`
				ExtraData   interface{} `json:"extra_data"`
				Forecaster  string      `json:"forecaster"`
				BlockHeight string      `json:"block_height"`
				ForecastElements []struct{
					Inferer	 string      `json:"inferer"`
					Value	 string      `json:"value"`
				} `json:"forecast_elements"`
			} `json:"forecast"`
			Inference struct {
				Proof       string      `json:"proof"`
				Value       string      `json:"value"`
				Inferer     string      `json:"inferer"`
				TopicID     string      `json:"topic_id"`
				ExtraData   interface{} `json:"extra_data"`
				BlockHeight string      `json:"block_height"`
			} `json:"inference"`
		} `json:"inferenceForecastsBundle"`
		InferencesForecastsBundleSignature string `json:"inferencesForecastsBundleSignature"`
	} `json:"worker_data_bundles"`
}


