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

func IsFutures(ticker string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(ticker))
	if len(trimmed) < 4 {
		return false
	}

	suffix := trimmed[len(trimmed)-3:]
	monthCode := suffix[0]
	yearSuffix := suffix[1:]

	if monthCode != 'F' && monthCode != 'G' && monthCode != 'H' && monthCode != 'J' && monthCode != 'K' && monthCode != 'M' && monthCode != 'N' && monthCode != 'Q' && monthCode != 'U' && monthCode != 'V' && monthCode != 'X' && monthCode != 'Z' {
		return false
	}

	if yearSuffix[0] < '0' || yearSuffix[0] > '9' || yearSuffix[1] < '0' || yearSuffix[1] > '9' {
		return false
	}

	return true
}
