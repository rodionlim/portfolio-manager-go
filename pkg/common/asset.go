package common

import (
	"strings"
)

func IsSSB(ticker string) bool {
	return strings.HasPrefix(ticker, "SB") && len(ticker) == 7
}

func IsSgTBill(ticker string) bool {
	return (strings.HasPrefix(ticker, "BS") || strings.HasPrefix(ticker, "BY")) && len(ticker) == 8
}
