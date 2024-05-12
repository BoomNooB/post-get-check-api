package model

type RequestBroadcastTxn struct {
	Symbol    string `json:"symbol"`
	Price     uint64 `json:"price"`
	TimeStamp uint64 `json:"timestamp"`
}

type RequestTxnStatusCheck struct {
	TXStatus string `json:"tx_status"`
}

type Response struct {
	TXHash string `json:"tx_hash"`
}
