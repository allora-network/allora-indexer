package types



type Tx struct {
	Body struct {
		Messages                    []Message `json:"messages,omitempty"`
		Memo                        string `json:"memo,omitempty"`
		TimeoutHeight               string `json:"timeout_height,omitempty"`
		ExtensionOptions            []any  `json:"extension_options,omitempty"`
		NonCriticalExtensionOptions []any  `json:"non_critical_extension_options,omitempty"`
	} `json:"body,omitempty"`
	AuthInfo struct {
		SignerInfos []struct {
			PublicKey struct {
				Type string `json:"@type,omitempty"`
				Key  string `json:"key,omitempty"`
			} `json:"public_key,omitempty"`
			ModeInfo struct {
				Single struct {
					Mode string `json:"mode,omitempty"`
				} `json:"single,omitempty"`
			} `json:"mode_info,omitempty"`
			Sequence string `json:"sequence,omitempty"`
		} `json:"signer_infos,omitempty"`
		Fee struct {
			Amount []struct {
				Denom  string `json:"denom,omitempty"`
				Amount string `json:"amount,omitempty"`
			} `json:"amount,omitempty"`
			GasLimit string `json:"gas_limit,omitempty"`
			Payer    string `json:"payer,omitempty"`
			Granter  string `json:"granter,omitempty"`
		} `json:"fee,omitempty"`
		Tip any `json:"tip,omitempty"`
	} `json:"auth_info,omitempty"`
	Signatures []string `json:"signatures,omitempty"`
}

type Message map[string]interface{}

type MsgSend struct {
	Type string `json:"@type,omitempty"`
	FromAddress string `json:"from_address,omitempty"`
	ToAddress string `json:"to_address,omitempty"`
	Amount []struct {
		Denom  string `json:"denom,omitempty"`
		Amount string `json:"amount,omitempty"`
	} `json:"amount,omitempty"`
}
type MsgFundTopic struct {
	Type string `json:"@type,omitempty"`
	Sender string `json:"sender,omitempty"`
	TopicID string `json:"topic_id,omitempty"`
	Amount string `json:"amount,omitempty"`
}


type MsgRegister struct {
	Type string `json:"@type,omitempty"`
	Sender string `json:"sender,omitempty"`
	TopicID string `json:"topic_id,omitempty"`
	Owner string `json:"owner,omitempty"`
	LibP2pKey string `json:"lib_p2p_key,omitempty"`
	MultiAddress string `json:"multi_address,omitempty"`
	IsReputer bool `json:"is_reputer,omitempty"`
}



// type Tx struct {
// 	Body struct {
// 		Messages []struct {
// 			Type             string `json:"@type,omitempty"`

// 			// MsgMsgRegister
// 			Sender           string `json:"sender,omitempty"`
// 			Owner            string `json:"owner,omitempty"`
// 			LibP2pKey        string `json:"lib_p2p_key,omitempty"`
// 			MultiAddress     string `json:"multi_address,omitempty"`
// 			TopicId          string `json:"topic_id,omitempty"`
// 			IsReputer        bool `json:"is_reputer,omitempty"`

// 			//MsgCreateNewTopic
// 			Creator	    string `json:"creator,omitempty"`
// 			Metadata	string `json:"metadata,omitempty"`
// 			LossLogic	string `json:"loss_logic,omitempty"`
// 			LossMethod	string `json:"loss_method,omitempty"`
// 			InferenceLogic	string `json:"inference_logic,omitempty"`
// 			InferenceMethod	string `json:"inference_method,omitempty"`
// 			EpochLength	    string `json:"epoch_length,omitempty"`
// 			GroundTruth_lag	string `json:"ground_truth_lag,omitempty"`
// 			DefaultArg	    string `json:"default_arg,omitempty"`
// 			PNorm	        string `json:"pnorm,omitempty"`
// 			AlphaRegret	    string `json:"alpha_regret,omitempty"`
// 			PrewardReputer	string `json:"preward_reputer,omitempty"`
// 			PrewardInference string `json:"preward_inference,omitempty"`
// 			PrewardForecast	 string `json:"preward_forecast,omitempty"`
// 			FTolerance	     string `json:"f_tolerance,omitempty"`

// 			// MsgInsertBulkWorkerPayload
// 			// Sender	string `json:"sender,omitempty"`
// 			Nonce struct {
// 				BlockHeight	string `json:"block_height,omitempty"`
// 			} `json:"nonce,omitempty"`
// 			// TopicId	string `json:"topic_id,omitempty"`
// 			WorkerDataBundles []struct {
// 				Worker	string `json:"worker,omitempty"`
// 				InferenceForecastsBundle struct {
// 					Inference struct {
// 						TopicId	string `json:"topic_id,omitempty"`
// 						BlockHeight	string `json:"block_height,omitempty"`
// 						Inferer	string `json:"inferer,omitempty"`
// 						Value	string `json:"value,omitempty"`
// 						ExtraData	string `json:"extra_data,omitempty"`
// 						Proof	string `json:"proof,omitempty"`
// 					} `json:"inference,omitempty"`
// 					Forecast struct {
// 						TopicId	string `json:"topic_id,omitempty"`
// 						BlockHeight	string `json:"block_height,omitempty"`
// 						Forecaster	string `json:"forecaster,omitempty"`
// 						ForecastElements	[]string `json:"forecast_elements,omitempty"`
// 						ExtraData	string `json:"extra_data,omitempty"`
// 					} `json:"forecast,omitempty"`
// 				} `json:"inference_forecasts_bundle,omitempty"`
// 				InferencesForecastsBundleSignature	string `json:"inferences_forecasts_bundle_signature,omitempty"`
// 				Pubkey	string `json:"pubkey,omitempty"`
// 			} `json:"worker_data_bundles,omitempty"`

// 			// emissions.v1.MsgAddStake
// 			// Sender	string `json:"sender,omitempty"`
// 			// TopicId	string `json:"topic_id,omitempty"`
// 			Amount	string `json:"amount,omitempty"`

// 			// DelegatorAddress string `json:"delegator_address,omitempty"`
// 			// ValidatorAddress string `json:"validator_address,omitempty"`
// 			// Amount           struct {
// 			// 	Denom  string `json:"denom,omitempty"`
// 			// 	Amount string `json:"amount,omitempty"`
// 			// } `json:"amount,omitempty"`
// 		} `json:"messages,omitempty"`
// 		Memo                        string `json:"memo,omitempty"`
// 		TimeoutHeight               string `json:"timeout_height,omitempty"`
// 		ExtensionOptions            []any  `json:"extension_options,omitempty"`
// 		NonCriticalExtensionOptions []any  `json:"non_critical_extension_options,omitempty"`
// 	} `json:"body,omitempty"`
// 	AuthInfo struct {
// 		SignerInfos []struct {
// 			PublicKey struct {
// 				Type string `json:"@type,omitempty"`
// 				Key  string `json:"key,omitempty"`
// 			} `json:"public_key,omitempty"`
// 			ModeInfo struct {
// 				Single struct {
// 					Mode string `json:"mode,omitempty"`
// 				} `json:"single,omitempty"`
// 			} `json:"mode_info,omitempty"`
// 			Sequence string `json:"sequence,omitempty"`
// 		} `json:"signer_infos,omitempty"`
// 		Fee struct {
// 			Amount []struct {
// 				Denom  string `json:"denom,omitempty"`
// 				Amount string `json:"amount,omitempty"`
// 			} `json:"amount,omitempty"`
// 			GasLimit string `json:"gas_limit,omitempty"`
// 			Payer    string `json:"payer,omitempty"`
// 			Granter  string `json:"granter,omitempty"`
// 		} `json:"fee,omitempty"`
// 		Tip any `json:"tip,omitempty"`
// 	} `json:"auth_info,omitempty"`
// 	Signatures []string `json:"signatures,omitempty"`
// }
