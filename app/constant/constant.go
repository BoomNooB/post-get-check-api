package constant

const (
	// txn status
	Confirmed    = "CONFIRMED"
	Failed       = "FAILED"
	Pending      = "PENDING"
	DoesNotExist = "DNE"

	// header
	XReqID = "X-Request-ID"

	// content type for req
	ContentTypeJSON = "application/json"

	// http method
	POST = "POST"

	// TXN_HASH
	TXNHash   = "tx_hash"
	TXNStatus = "tx_status"

	// retry
	RetryCount = "retry count"
)
