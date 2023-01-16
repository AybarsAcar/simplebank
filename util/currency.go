package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	AUD = "AUD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, AUD, CAD:
		return true
	}

	return false
}
