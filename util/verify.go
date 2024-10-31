package util

import "regexp"

var (
	TronAddressRegex = regexp.MustCompile(`^41[0-9a-fA-F]{40}$`)
	EthAddressRegex  = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	BtcAddressRegex  = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
)

func VerifyTranAddress(address string) bool {
	return TronAddressRegex.MatchString(address)
}

func VerifyEthAddress(address string) bool {
	return EthAddressRegex.MatchString(address)
}

func VerifyBtcAddress(address string) bool {
	return BtcAddressRegex.MatchString(address)
}

func VerifyAddress(address, chain string) bool {
	switch chain {
	case "Tron":
		return VerifyTranAddress(address)
	default:
		return VerifyEthAddress(address)
	}
}

func GetChainByAddress(address string) string {
	if VerifyTranAddress(address) {
		return "Tron"
	} else {
		return "Eth"
	}
}
