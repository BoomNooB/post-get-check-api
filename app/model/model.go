package model

type RequestBodyBroadcastTxn struct {
	Symbol    string `json:"symbol" validate:"required"`
	Price     uint64 `json:"price" validate:"required"`
	TimeStamp uint64 `json:"timestamp" validate:"required"`
}

type RequestTxnStatusCheck struct {
	TXStatus string `json:"tx_status"`
}

type RequestBodyTxnStatusCheckExt struct {
	TXStatus string `json:"tx_status" validate:"required"`
}

type ResponseFromBroadcastTxn struct {
	TXHash string `json:"tx_hash"`
}

type ResponseExternal struct {
	Message  string `json:"msg"`
	TXStatus string `json:"tx_status,omitempty"`
	TXHash   string `json:"tx_hash,omitempty"`
}
