package types

type Tx struct {
	Body struct {
		Messages []struct {
			Type             string `json:"@type,omitempty"`
			DelegatorAddress string `json:"delegator_address,omitempty"`
			ValidatorAddress string `json:"validator_address,omitempty"`
			Amount           struct {
				Denom  string `json:"denom,omitempty"`
				Amount string `json:"amount,omitempty"`
			} `json:"amount,omitempty"`
		} `json:"messages,omitempty"`
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
