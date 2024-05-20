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
