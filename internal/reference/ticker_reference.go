package reference

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	if validate == nil {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterValidation("asset_class", validateAssetClass)
		validate.RegisterValidation("asset_sub_class", validateAssetSubClass)
	}
}

type TickerReference struct {
	ID                string  `json:"id" validate:"required"`
	YahooTicker       string  `json:"yahoo_ticker"`
	GoogleTicker      string  `json:"google_ticker"`
	DividendsSgTicker string  `json:"dividends_sg_ticker"`
	AssetClass        string  `json:"asset_class" validate:"required,asset_class"`
	AssetSubClass     string  `json:"asset_sub_class" validate:"asset_sub_class"`
	CouponRate        float64 `json:"coupon_rate"`
	MaturityDate      string  `json:"maturity_date"`
	StrikePrice       float64 `json:"strike_price"`
	CallPut           string  `json:"call_put" validate:"oneof=call put"`
}

// Supported asset classes
const (
	AssetClassFX          = "fx"
	AssetClassEquities    = "eq"
	AssetClassCrypto      = "crypto"
	AssetClassCommodities = "cmdty"
	AssetClassCash        = "cash"
	AssetClassBonds       = "bond"
)

// Supported asset sub classes
const (
	AssetSubClassStock  = "stock"
	AssetSubClassETF    = "etf"
	AssetSubClassReit   = "reit"
	AssetSubClassOption = "option"
	AssetSubClassFuture = "future"
	AssetSubClassCash   = "cash"
)

// NewTickerReference creates a new TickerReference instance.
func NewTickerReference(id, yahooTicker, googleTicker, dividendsSgTicker, assetClass, assetSubClass string, couponRate, strikePrice float64, maturityDate, callPut string) (TickerReference, error) {
	ref := TickerReference{
		ID:                id,
		YahooTicker:       yahooTicker,
		GoogleTicker:      googleTicker,
		DividendsSgTicker: dividendsSgTicker,
		AssetClass:        assetClass,
		AssetSubClass:     assetSubClass,
		CouponRate:        couponRate,
		MaturityDate:      maturityDate,
		StrikePrice:       strikePrice,
		CallPut:           callPut,
	}

	err := validate.Struct(ref)
	return ref, err
}

func validateAssetClass(fl validator.FieldLevel) bool {
	ac := fl.Field().String()

	switch ac {
	case AssetClassFX, AssetClassEquities, AssetClassCrypto, AssetClassCommodities, AssetClassCash, AssetClassBonds:
		return true
	default:
		return false
	}
}

func validateAssetSubClass(fl validator.FieldLevel) bool {
	asc := fl.Field().String()

	switch asc {
	case AssetSubClassStock, AssetSubClassETF, AssetSubClassReit, AssetSubClassOption, AssetSubClassFuture, AssetSubClassCash:
		return true
	default:
		return false
	}
}
