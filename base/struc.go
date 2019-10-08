package base

type Struct struct {
	ReqType string `json:"type"`
	Content []byte `json:"content"`
	SessionId string `json:"sessionId"`
}
