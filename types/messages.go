package types

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
				ExtraData   interface{} `json:"extra_data,omitempty"`
				Forecaster  string      `json:"forecaster"`
				BlockHeight string      `json:"block_height"`
				ForecastElements []struct{
					Inferer	 string      `json:"inferer"`
					Value	 string      `json:"value"`
				} `json:"forecast_elements"`
			} `json:"forecast,omitempty"`
			Inference struct {
				Proof       string      `json:"proof"`
				Value       string      `json:"value"`
				Inferer     string      `json:"inferer"`
				TopicID     string      `json:"topic_id"`
				ExtraData   interface{} `json:"extra_data,omitempty"`
				BlockHeight string      `json:"block_height"`
			} `json:"inference,omitempty"`
		} `json:"inference_forecasts_bundle"`
		InferencesForecastsBundleSignature string `json:"inferences_forecasts_bundle_signature"`
	} `json:"worker_data_bundles"`
}

type InsertBulkReputerPayload struct {
	Type                  string `json:"@type"`
	Sender                string `json:"sender"`
	TopicID               string `json:"topic_id"`
	ReputerRequestNonce   struct {
		WorkerNonce struct {
			BlockHeight string `json:"block_height"`
		} `json:"worker_nonce"`
		ReputerNonce struct {
			BlockHeight string `json:"block_height"`
		} `json:"reputer_nonce"`
	} `json:"reputer_request_nonce"`
	ReputerValueBundles   []struct {
		Pubkey                      string `json:"pubkey"`
		Signature                   string `json:"signature"`
		ValueBundle                 struct {
			Reputer                 string `json:"reputer"`
			TopicID                 string `json:"topic_id"`
			ExtraData               interface{} `json:"extra_data,omitempty"`
			NaiveValue              string `json:"naive_value"`
			CombinedValue           string `json:"combined_value"`
			InfererValues           []struct {
				Value               string `json:"value"`
				Worker              string `json:"worker"`
			} `json:"inferer_values"`
			ForecasterValues        []struct {
				Value               string `json:"value"`
				Worker              string `json:"worker"`
			} `json:"forecaster_values"`
			ReputerRequestNonce     struct {
				WorkerNonce struct {
					BlockHeight string `json:"block_height"`
				} `json:"worker_nonce"`
				ReputerNonce struct {
					BlockHeight string `json:"block_height"`
				} `json:"reputer_nonce"`
			} `json:"reputer_request_nonce"`
			OneOutInfererValues     []struct {
				Value               string `json:"value"`
				Worker              string `json:"worker"`
			} `json:"one_out_inferer_values"`
			OneInForecasterValues   []struct {
				Value               string `json:"value"`
				Worker              string `json:"worker"`
			} `json:"one_in_forecaster_values"`
			OneOutForecasterValues  []struct {
				Value               string `json:"value"`
				Worker              string `json:"worker"`
			} `json:"one_out_forecaster_values"`
		} `json:"value_bundle"`
	} `json:"reputer_value_bundles"`
}
