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
	NasdaqTicker      string  `json:"nasdaq_ticker" yaml:"nasdaq_ticker" validate:"uppercase"`
	BarchartTicker    string  `json:"barchart_ticker" yaml:"barchart_ticker" validate:"uppercase"`
	AssetClass        string  `json:"asset_class" yaml:"asset_class" validate:"required,asset_class"`
	AssetSubClass     string  `json:"asset_sub_class" yaml:"asset_sub_class" validate:"asset_sub_class"`
	Category          string  `json:"category" yaml:"category" validate:"category"`
	SubCategory       string  `json:"sub_category" yaml:"sub_category"`
	Ccy               string  `json:"ccy" yaml:"ccy" validate:"required,uppercase"`
	Domicile          string  `json:"domicile" yaml:"domicile" validate:"required,uppercase"`
	CouponRate        float64 `json:"coupon_rate" yaml:"coupon_rate"`
	MaturityDate      string  `json:"maturity_date" yaml:"maturity_date" validate:"omitempty,datetime=2006-01-02"`
	StrikePrice       float64 `json:"strike_price" yaml:"strike_price"`
	CallPut           string  `json:"call_put" yaml:"call_put" validate:"oneof=call put"`
}

// TickerReferenceWithSGXMapped extends TickerReference with SGX-compatible category mapping
type TickerReferenceWithSGXMapped struct {
	TickerReference
	CategorySGX string `json:"category_sgx" yaml:"category_sgx"`
}

// categoryToSGXMap maps internal category names to SGX sector names
var categoryToSGXMap = map[string]string{
	CategoryConsumerCyclicals:    "Consumer Cyclicals",
	CategoryConsumerNonCyclicals: "Consumer Non-Cyclicals",
	CategoryEnergy:               "Energy/Oil & Gas",
	CategoryFinance:              "Financial Services",
	CategoryHealthcare:           "Healthcare",
	CategoryIndustrials:          "Industrials",
	CategoryMaterials:            "Materials & Resources",
	CategoryRealEstate:           "Real Estate (excl. REITs)",
	CategoryREITs:                "REITs",
	CategoryTechnology:           "Technology (Hardware/Software)",
	CategoryTelecommunications:   "Telcos",
	CategoryUtilities:            "Utilities",
}

// MapCategoryToSGX converts an internal category to SGX sector name
// Returns the original category if no mapping exists
func MapCategoryToSGX(category string) string {
	if sgxCategory, ok := categoryToSGXMap[category]; ok {
		return sgxCategory
	}
	return category // Return original if no mapping found
}

// ToSGXMapped converts a TickerReference to TickerReferenceWithSGXMapped
func (t TickerReference) ToSGXMapped() TickerReferenceWithSGXMapped {
	return TickerReferenceWithSGXMapped{
		TickerReference: t,
		CategorySGX:     MapCategoryToSGX(t.Category),
	}
}

// Supported asset classes
const (
	AssetClassBonds       = "bond"
	AssetClassCash        = "cash"
	AssetClassCommodities = "cmdty"
	AssetClassCrypto      = "crypto"
	AssetClassEquities    = "eq"
	AssetClassFX          = "fx"
)

// Supported asset sub classes
const (
	AssetSubClassBond   = "bond"
	AssetSubClassCash   = "cash"
	AssetSubClassETF    = "etf"
	AssetSubClassFuture = "future"
	AssetSubClassGovies = "govies"
	AssetSubClassOption = "option"
	AssetSubClassReit   = "reit"
	AssetSubClassSpot   = "spot"
	AssetSubClassStock  = "stock"
)

// Supported categories
const (
	CategoryAgriculture          = "agriculture"
	CategoryConsumerCyclicals    = "consumercyclicals"
	CategoryConsumerNonCyclicals = "consumernoncyclicals"
	CategoryCrypto               = "crypto"
	CategoryEnergy               = "energy"
	CategoryFinance              = "finance"
	CategoryFuneral              = "funeral"
	CategoryHealthcare           = "healthcare"
	CategoryIndustrials          = "industrials"
	CategoryMaterials            = "materials"  // Matericals and Resources
	CategoryRealEstate           = "realestate" // Excludes REITs
	CategoryREITs                = "reits"
	CategoryTelecommunications   = "telecommunications"
	CategoryTechnology           = "technology"
	CategoryUtilities            = "utilities"
)

// NewTickerReference creates a new TickerReference instance.
func NewTickerReference(id, name, underlyingTicker, yahooTicker, googleTicker, dividendsSgTicker, nasdaqTicker, barchartTicker, assetClass, assetSubClass, category, subcategory, ccy, domicile string, couponRate, strikePrice float64, maturityDate, callPut string) (*TickerReference, error) {
	ref := TickerReference{
		ID:                id,
		Name:              name,
		UnderlyingTicker:  underlyingTicker,
		YahooTicker:       yahooTicker,
		GoogleTicker:      googleTicker,
		DividendsSgTicker: dividendsSgTicker,
		NasdaqTicker:      nasdaqTicker,
		BarchartTicker:    barchartTicker,
		AssetClass:        assetClass,
		AssetSubClass:     assetSubClass,
		Category:          category,
		SubCategory:       subcategory,
		Ccy:               ccy,
		Domicile:          domicile,
		CouponRate:        couponRate,
		MaturityDate:      maturityDate, // YYYY-MM-DD
		StrikePrice:       strikePrice,
		CallPut:           callPut,
	}

	err := validate.Struct(ref)
	return &ref, err
}

func validateAssetClass(fl validator.FieldLevel) bool {
	ac := fl.Field().String()

	switch ac {
	case AssetClassBonds, AssetClassCash, AssetClassCommodities, AssetClassCrypto, AssetClassEquities, AssetClassFX:
		return true
	default:
		return false
	}
}

func validateAssetSubClass(fl validator.FieldLevel) bool {
	asc := fl.Field().String()

	switch asc {
	case AssetSubClassCash, AssetSubClassETF, AssetSubClassFuture, AssetSubClassGovies, AssetSubClassOption, AssetSubClassReit, AssetSubClassSpot, AssetSubClassStock:
		return true
	default:
		return false
	}
}

func validateCategory(fl validator.FieldLevel) bool {
	cat := fl.Field().String()

	switch cat {
	case CategoryConsumerCyclicals, CategoryConsumerNonCyclicals, CategoryCrypto, CategoryEnergy, CategoryFinance, CategoryFuneral, CategoryHealthcare, CategoryIndustrials, CategoryMaterials, CategoryRealEstate, CategoryREITs, CategoryTelecommunications, CategoryTechnology, CategoryUtilities, CategoryAgriculture:
		return true
	default:
		return false
	}
}
