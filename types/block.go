package types

import "time"

type BlockInfo struct {
	BlockID struct {
		Hash          string `json:"hash,omitempty"`
		PartSetHeader struct {
			Total int    `json:"total,omitempty"`
			Hash  string `json:"hash,omitempty"`
		} `json:"part_set_header,omitempty"`
	} `json:"block_id,omitempty"`
	Block struct {
		Header struct {
			Version struct {
				Block string `json:"block,omitempty"`
			} `json:"version,omitempty"`
			ChainID     string    `json:"chain_id,omitempty"`
			Height      string    `json:"height,omitempty"`
			Time        time.Time `json:"time,omitempty"`
			LastBlockID struct {
				Hash          string `json:"hash,omitempty"`
				PartSetHeader struct {
					Total int    `json:"total,omitempty"`
					Hash  string `json:"hash,omitempty"`
				} `json:"part_set_header,omitempty"`
			} `json:"last_block_id,omitempty"`
			LastCommitHash     string `json:"last_commit_hash,omitempty"`
			DataHash           string `json:"data_hash,omitempty"`
			ValidatorsHash     string `json:"validators_hash,omitempty"`
			NextValidatorsHash string `json:"next_validators_hash,omitempty"`
			ConsensusHash      string `json:"consensus_hash,omitempty"`
			AppHash            string `json:"app_hash,omitempty"`
			LastResultsHash    string `json:"last_results_hash,omitempty"`
			EvidenceHash       string `json:"evidence_hash,omitempty"`
			ProposerAddress    string `json:"proposer_address,omitempty"`
		} `json:"header,omitempty"`
		Data struct {
		} `json:"data,omitempty"`
		Evidence struct {
		} `json:"evidence,omitempty"`
		LastCommit struct {
			Height  string `json:"height,omitempty"`
			BlockID struct {
				Hash          string `json:"hash,omitempty"`
				PartSetHeader struct {
					Total int    `json:"total,omitempty"`
					Hash  string `json:"hash,omitempty"`
				} `json:"part_set_header,omitempty"`
			} `json:"block_id,omitempty"`
			Signatures []struct {
				BlockIDFlag      int       `json:"block_id_flag,omitempty"`
				ValidatorAddress string    `json:"validator_address,omitempty"`
				Timestamp        time.Time `json:"timestamp,omitempty"`
				Signature        string    `json:"signature,omitempty"`
			} `json:"signatures,omitempty"`
		} `json:"last_commit,omitempty"`
	} `json:"block,omitempty"`
	SdkBlock struct {
		Header struct {
			Version struct {
				Block string `json:"block,omitempty"`
			} `json:"version,omitempty"`
			ChainID     string    `json:"chain_id,omitempty"`
			Height      string    `json:"height,omitempty"`
			Time        time.Time `json:"time,omitempty"`
			LastBlockID struct {
				Hash          string `json:"hash,omitempty"`
				PartSetHeader struct {
					Total int    `json:"total,omitempty"`
					Hash  string `json:"hash,omitempty"`
				} `json:"part_set_header,omitempty"`
			} `json:"last_block_id,omitempty"`
			LastCommitHash     string `json:"last_commit_hash,omitempty"`
			DataHash           string `json:"data_hash,omitempty"`
			ValidatorsHash     string `json:"validators_hash,omitempty"`
			NextValidatorsHash string `json:"next_validators_hash,omitempty"`
			ConsensusHash      string `json:"consensus_hash,omitempty"`
			AppHash            string `json:"app_hash,omitempty"`
			LastResultsHash    string `json:"last_results_hash,omitempty"`
			EvidenceHash       string `json:"evidence_hash,omitempty"`
			ProposerAddress    string `json:"proposer_address,omitempty"`
		} `json:"header,omitempty"`
		Data struct {
		} `json:"data,omitempty"`
		Evidence struct {
		} `json:"evidence,omitempty"`
		LastCommit struct {
			Height  string `json:"height,omitempty"`
			BlockID struct {
				Hash          string `json:"hash,omitempty"`
				PartSetHeader struct {
					Total int    `json:"total,omitempty"`
					Hash  string `json:"hash,omitempty"`
				} `json:"part_set_header,omitempty"`
			} `json:"block_id,omitempty"`
			Signatures []struct {
				BlockIDFlag      int       `json:"block_id_flag,omitempty"`
				ValidatorAddress string    `json:"validator_address,omitempty"`
				Timestamp        time.Time `json:"timestamp,omitempty"`
				Signature        string    `json:"signature,omitempty"`
			} `json:"signatures,omitempty"`
		} `json:"last_commit,omitempty"`
	} `json:"sdk_block,omitempty"`
}
