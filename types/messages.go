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
