package types

type MsgCreateNewTopic struct {
	Type             string `json:"@type,omitempty"` // Optional to be able to use to unmarshal topic by id
	TopicID          string `json:"id,omitempty"`    // Optional to be able to use to unmarshal topic by id
	Creator          string `json:"creator"`
	Metadata         string `json:"metadata"`
	LossLogic        string `json:"loss_logic"`
	LossMethod       string `json:"loss_method"`
	InferenceLogic   string `json:"inference_logic"`
	InferenceMethod  string `json:"inference_method"`
	EpochLength      string `json:"epoch_length"`
	GroundTruthLag   string `json:"ground_truth_lag"`
	DefaultArg       string `json:"default_arg"`
	Pnorm            string `json:"pnorm"`
	AlphaRegret      string `json:"alpha_regret"`
	PrewardReputer   string `json:"preward_reputer"`
	PrewardInference string `json:"preward_inference"`
	PrewardForecast  string `json:"preward_forecast"`
	FTolerance       string `json:"f_tolerance"`
	AllowNegative    bool   `json:"allow_negative"`
}

type MsgInsertWorkerPayload struct {
	Type             string `json:"@type"`
	Sender           string `json:"sender"`
	WorkerDataBundle struct {
		Worker string `json:"worker"`
		Nonce  struct {
			BlockHeight string `json:"block_height"`
		} `json:"nonce"`
		TopicID                  string `json:"topic_id"`
		InferenceForecastsBundle struct {
			Forecast struct {
				TopicID          string      `json:"topic_id"`
				ExtraData        interface{} `json:"extra_data,omitempty"`
				Forecaster       string      `json:"forecaster"`
				BlockHeight      string      `json:"block_height"`
				ForecastElements []struct {
					Inferer string `json:"inferer"`
					Value   string `json:"value"`
				} `json:"forecast_elements,omitempty"`
			} `json:"forecast,omitempty"`
			Inference struct {
				Value       string      `json:"value"`
				Inferer     string      `json:"inferer"`
				TopicID     string      `json:"topic_id"`
				ExtraData   interface{} `json:"extra_data,omitempty"`
				BlockHeight string      `json:"block_height"`
				Proof       string      `json:"proof"`
			} `json:"inference,omitempty"`
		} `json:"inference_forecasts_bundle"`
		InferencesForecastsBundleSignature string `json:"inferences_forecasts_bundle_signature"`
		PubKey                             string `json:"pubkey"`
	} `json:"worker_data_bundle"`
}

type MsgValueBundle struct {
	Reputer       string      `json:"reputer"`
	TopicID       string      `json:"topic_id"`
	ExtraData     interface{} `json:"extra_data,omitempty"`
	NaiveValue    string      `json:"naive_value"`
	CombinedValue string      `json:"combined_value"`
	InfererValues []struct {
		Value  string `json:"value"`
		Worker string `json:"worker"`
	} `json:"inferer_values"`
	ForecasterValues []struct {
		Value  string `json:"value"`
		Worker string `json:"worker"`
	} `json:"forecaster_values"`
	ReputerRequestNonce struct {
		ReputerNonce struct {
			BlockHeight string `json:"block_height"`
		} `json:"reputer_nonce"`
	} `json:"reputer_request_nonce"`
	OneOutInfererValues []struct {
		Value  string `json:"value"`
		Worker string `json:"worker"`
	} `json:"one_out_inferer_values"`
	OneInForecasterValues []struct {
		Value  string `json:"value"`
		Worker string `json:"worker"`
	} `json:"one_in_forecaster_values"`
	OneOutForecasterValues []struct {
		Value  string `json:"value"`
		Worker string `json:"worker"`
	} `json:"one_out_forecaster_values"`
	OneOutInfererForecasterValues []struct {
		Forecaster          string `json:"forecaster"`
		OneOutInfererValues []struct {
			Value  string `json:"value"`
			Worker string `json:"worker"`
		} `json:"one_out_inferer_values"`
	} `json:"one_out_inferer_forecaster_values"`
}

type MsgInsertReputerPayload struct {
	Type               string `json:"@type"`
	Sender             string `json:"sender"`
	ReputerValueBundle struct {
		Pubkey      string         `json:"pubkey"`
		Signature   string         `json:"signature"`
		ValueBundle MsgValueBundle `json:"value_bundle"`
	} `json:"reputer_value_bundle"`
}

type MsgSend struct {
	Type        string `json:"@type,omitempty"`
	FromAddress string `json:"from_address,omitempty"`
	ToAddress   string `json:"to_address,omitempty"`
	Amount      []struct {
		Denom  string `json:"denom,omitempty"`
		Amount string `json:"amount,omitempty"`
	} `json:"amount,omitempty"`
}
type MsgFundTopic struct {
	Type    string `json:"@type,omitempty"`
	Sender  string `json:"sender,omitempty"`
	TopicID string `json:"topic_id,omitempty"`
	Amount  string `json:"amount,omitempty"`
}

type MsgRegister struct {
	Type         string `json:"@type,omitempty"`
	Sender       string `json:"sender,omitempty"`
	TopicID      string `json:"topic_id,omitempty"`
	Owner        string `json:"owner,omitempty"`
	LibP2pKey    string `json:"lib_p2p_key,omitempty"`
	MultiAddress string `json:"multi_address,omitempty"`
	IsReputer    bool   `json:"is_reputer,omitempty"`
}
