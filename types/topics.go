package types


type Topic struct {
	TopicID           string `json:"id"`
	Creator           string `json:"creator"`
	Metadata          string `json:"metadata"`
	LossLogic         string `json:"loss_logic"`
	LossMethod        string `json:"loss_method"`
	InferenceLogic    string `json:"inference_logic"`
	InferenceMethod   string `json:"inference_method"`
	EpochLastEnded    string `json:"epoch_last_ended"`
	EpochLength       string `json:"epoch_length"`
	GroundTruthLag    string `json:"ground_truth_lag"`
	DefaultArg        string `json:"default_arg"`
	Pnorm             string `json:"pnorm"`
	AlphaRegret       string `json:"alpha_regret"`
	PrewardReputer    string `json:"preward_reputer"`
	PrewardInference  string `json:"preward_inference"`
	PrewardForecast   string `json:"preward_forecast"`
	FTolerance        string `json:"f_tolerance"`
}
