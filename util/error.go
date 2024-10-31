package util

type QYError struct {
	Code int `json:"code"`
	// Message string "json: message"
}

var QYErrorMap = map[int]string{
	// system
	10000: "fail",
	10001: "params error",
	10002: "verify signature error",
	10003: "wallet address error",
	10004: "email exist",
	10005: "name exist",
	10006: "wallet exist",
	10007: "wallet exist",
	10008: "bad sso query",
	10009: "token set error",
	10010: "no permission",
	10011: "balance insufficient",
	10012: "nonce  error",
	10013: "sign message error",
	10014: "has unfinished Withdraw",
	10015: "Transaction with the same hash was already imported.",
	10016: "treasure not found",
	10017: "treasure has unlock",
	10018: "open treasure fail",
	10019: "field conflict",
	10020: "bank trade data not found",
	10021: "upload file limit 50kb",
	10022: "not support upload type",
	10023: "upload file fail",
	10024: "keystore examine fail",
	10025: "upload land cover fail",
	10026: "already bind keystore",
	10027: "invalid phone number",
	10028: "invalid sms code",
	10029: "already grabbed",
	10030: "has open this red packet",
	10031: "your have bind mobile",
	10032: "this mobile has been bind",
	10033: "your have bind email",
	10034: "sms send limit",
	10035: "not bind wallet",
	10036: "has bind wallet",
	10037: "invalid app",
	10038: "no right to operate",
	10039: "pve adventure has been started",
	10040: "pve need 4 apostles",
	10041: "stage not exist",
	10042: "need choose one card",
	10043: "invalid card",
	30001: "upgrade in progress",
	30002: "the building has reached the highest level",
	30003: "the building upgrade complete",
	10404: "record not found",
	20001: "itering login error",
	20002: "not have any data",
	20003: "Can only request 8 times in 60 seconds",
	20004: "sso token is expired",
	77777: "json Unmarshal Error.",
	88888: "callback error,try again",
	88889: "tx exist",
	99999: "need login",
}

func (qyErr *QYError) GetCode() string {
	return QYErrorMap[qyErr.Code]
}

func (qyErr QYError) Error() string {
	return QYErrorMap[qyErr.Code]
}
