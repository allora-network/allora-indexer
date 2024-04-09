package types

type ConsensusParams struct {
	Params struct {
		Block struct {
			MaxBytes string `json:"max_bytes,omitempty"`
			MaxGas   string `json:"max_gas,omitempty"`
		} `json:"block,omitempty"`
		Evidence struct {
			MaxAgeNumBlocks string `json:"max_age_num_blocks,omitempty"`
			MaxAgeDuration  string `json:"max_age_duration,omitempty"`
			MaxBytes        string `json:"max_bytes,omitempty"`
		} `json:"evidence,omitempty"`
		Validator struct {
			PubKeyTypes []string `json:"pub_key_types,omitempty"`
		} `json:"validator,omitempty"`
		Version struct {
		} `json:"version,omitempty"`
		Abci struct {
		} `json:"abci,omitempty"`
	} `json:"params,omitempty"`
}
