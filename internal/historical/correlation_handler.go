package historical

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/common"
	"time"
)

type CorrelationOptions struct {
	Frequency             *string  `json:"frequency"`
	DateMethod            *string  `json:"date_method"`
	RollYears             *int     `json:"rollyears"`
	IntervalFrequency     *string  `json:"interval_frequency"`
	UsingExponent         *bool    `json:"using_exponent"`
	EwLookback            *int     `json:"ew_lookback"`
	MinPeriods            *int     `json:"min_periods"`
	FloorAtZero           *bool    `json:"floor_at_zero"`
	Clip                  *float64 `json:"clip"`
	Shrinkage             *float64 `json:"shrinkage"`
	ForwardFillPriceIndex *bool    `json:"forward_fill_price_index"`
	IndexCol              *int     `json:"index_col"`
}

type CorrelationRequest struct {
	Tickers []string           `json:"tickers"`
	From    string             `json:"from"` // YYYY-MM-DD
	To      string             `json:"to"`   // YYYY-MM-DD
	Options CorrelationOptions `json:"options"`
}

// HandleCalculateCorrelations handles the calculation of historical correlations
// @Summary Calculate historical correlations
// @Description Calculate correlation matrix using quantlib
// @Tags historical
// @Accept json
// @Produce json
// @Param request body CorrelationRequest true "Correlation Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/historical/correlation [post]
func HandleCalculateCorrelations(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CorrelationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.Tickers) < 2 {
			common.WriteJSONError(w, "at least two tickers are required", http.StatusBadRequest)
			return
		}

		fromTime, err := time.Parse("2006-01-02", req.From)
		if err != nil {
			common.WriteJSONError(w, "invalid from date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		toTime, err := time.Parse("2006-01-02", req.To)
		if err != nil {
			common.WriteJSONError(w, "invalid to date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		flags := buildFlags(req.Options)

		res, err := service.CalculateCorrelation(r.Context(), req.Tickers, fromTime.Unix(), toTime.Unix(), flags)
		if err != nil {
			common.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
}

func buildFlags(opts CorrelationOptions) []string {
	var flags []string
	if opts.Frequency != nil {
		flags = append(flags, "--frequency", *opts.Frequency)
	}
	if opts.DateMethod != nil {
		flags = append(flags, "--date-method", *opts.DateMethod)
	}
	if opts.RollYears != nil {
		flags = append(flags, "--rollyears", fmt.Sprintf("%d", *opts.RollYears))
	}
	if opts.IntervalFrequency != nil {
		flags = append(flags, "--interval-frequency", *opts.IntervalFrequency)
	}
	if opts.UsingExponent != nil {
		if *opts.UsingExponent {
			flags = append(flags, "--using-exponent")
		} else {
			flags = append(flags, "--no-using-exponent")
		}
	}
	if opts.EwLookback != nil {
		flags = append(flags, "--ew-lookback", fmt.Sprintf("%d", *opts.EwLookback))
	}
	if opts.MinPeriods != nil {
		flags = append(flags, "--min-periods", fmt.Sprintf("%d", *opts.MinPeriods))
	}
	if opts.FloorAtZero != nil {
		if *opts.FloorAtZero {
			flags = append(flags, "--floor-at-zero")
		} else {
			flags = append(flags, "--no-floor-at-zero")
		}
	}
	if opts.Clip != nil {
		flags = append(flags, "--clip", fmt.Sprintf("%f", *opts.Clip))
	}
	if opts.Shrinkage != nil {
		flags = append(flags, "--shrinkage", fmt.Sprintf("%f", *opts.Shrinkage))
	}
	if opts.ForwardFillPriceIndex != nil {
		if *opts.ForwardFillPriceIndex {
			flags = append(flags, "--forward-fill-price-index")
		} else {
			flags = append(flags, "--no-forward-fill-price-index")
		}
	}
	if opts.IndexCol != nil {
		flags = append(flags, "--index-col", fmt.Sprintf("%d", *opts.IndexCol))
	}
	return flags
}
