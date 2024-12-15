package common

import (
	"strings"
)

func IsSSB(ticker string) bool {
	return strings.HasPrefix(ticker, "SB") && len(ticker) == 7
}
