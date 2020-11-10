package log

type Message struct {
	ChainID     string `json:"chainID"`
	Level       string `json:"level"`
	Version     string `json:"version"`
	ServiceCode string `json:"serviceCode"`
	Time        string `json:"time"`
	Msg         string `json:"msg"`
}
