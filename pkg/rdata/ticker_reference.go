package rdata

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	if validate == nil {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterValidation("asset_class", validateAssetClass)
		validate.RegisterValidation("asset_sub_class", validateAssetSubClass)
		validate.RegisterValidation("category", validateCategory)
	}
}

type TickerReference struct {
	ID                string  `json:"id" yaml:"id" validate:"required,uppercase"`
	Name              string  `json:"name" yaml:"name" validate:"required"`
	UnderlyingTicker  string  `json:"underlying_ticker" yaml:"underlying_ticker" validate:"required,uppercase"`
	YahooTicker       string  `json:"yahoo_ticker" yaml:"yahoo_ticker" validate:"uppercase"`
	GoogleTicker      string  `json:"google_ticker" yaml:"google_ticker" validate:"uppercase"`
	DividendsSgTicker string  `json:"dividends_sg_ticker" yaml:"dividends_sg_ticker" validate:"uppercase"`
	AssetClass        string  `json:"asset_class" yaml:"asset_class" validate:"required,asset_class"`
	AssetSubClass     string  `json:"asset_sub_class" yaml:"asset_sub_class" validate:"asset_sub_class"`
	Category          string  `json:"category" yaml:"category" validate:"category"`
	SubCategory       string  `json:"sub_category" yaml:"sub_category"`
	Ccy               string  `json:"ccy" yaml:"ccy" validate:"required,uppercase"`
	Domicile          string  `json:"domicile" yaml:"domicile" validate:"required,uppercase"`
	CouponRate        float64 `json:"coupon_rate" yaml:"coupon_rate"`
	MaturityDate      string  `json:"maturity_date" yaml:"maturity_date"`
	StrikePrice       float64 `json:"strike_price" yaml:"strike_price"`
	CallPut           string  `json:"call_put" yaml:"call_put" validate:"oneof=call put"`
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
	AssetSubClassBond   = "bond"
	AssetSubClassGovies = "govies"
)

// Supported categories
const (
	CategoryTechnology    = "technology"
	CategoryFinance       = "finance"
	CategoryHealthcare    = "healthcare"
	CategoryEnergy        = "energy"
	CategoryRealEstate    = "realestate"
	CategoryConsumerGoods = "consumergoods"
	CategoryIndustrials   = "industrials"
	CategoryTravel        = "travel"
	CategoryFuneral       = "funeral"
)

// NewTickerReference creates a new TickerReference instance.
func NewTickerReference(id, name, underlyingTicker, yahooTicker, googleTicker, dividendsSgTicker, assetClass, assetSubClass, category, subcategory, ccy, domicile string, couponRate, strikePrice float64, maturityDate, callPut string) (*TickerReference, error) {
	ref := TickerReference{
		ID:                id,
		Name:              name,
		UnderlyingTicker:  underlyingTicker,
		YahooTicker:       yahooTicker,
		GoogleTicker:      googleTicker,
		DividendsSgTicker: dividendsSgTicker,
		AssetClass:        assetClass,
		AssetSubClass:     assetSubClass,
		Category:          category,
		SubCategory:       subcategory,
		Ccy:               ccy,
		Domicile:          domicile,
		CouponRate:        couponRate,
		MaturityDate:      maturityDate,
		StrikePrice:       strikePrice,
		CallPut:           callPut,
	}

	err := validate.Struct(ref)
	return &ref, err
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
	case AssetSubClassStock, AssetSubClassETF, AssetSubClassReit, AssetSubClassOption, AssetSubClassFuture, AssetSubClassCash, AssetSubClassGovies:
		return true
	default:
		return false
	}
}

func validateCategory(fl validator.FieldLevel) bool {
	cat := fl.Field().String()

	switch cat {
	case CategoryTechnology, CategoryFinance, CategoryConsumerGoods, CategoryEnergy, CategoryHealthcare, CategoryIndustrials, CategoryRealEstate:
		return true
	default:
		return false
	}
}