package reference

import (
	"encoding/json"
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/types"

	"github.com/google/uuid"
)

type ReferenceManager struct {
	db dal.Database
}

func NewReferenceManager(db dal.Database) (*ReferenceManager, error) {
	return &ReferenceManager{db: db}, nil
}

func (rm *ReferenceManager) AddTicker(ticker TickerReference) (string, error) {
	ticker.ID = uuid.New().String()
	data, err := json.Marshal(ticker)
	if err != nil {
		return "", err
	}
	err = rm.db.Put(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, ticker.ID), data)
	if err != nil {
		return "", err
	}
	return ticker.ID, nil
}

func (rm *ReferenceManager) UpdateTicker(ticker TickerReference) error {
	if ticker.ID == "" {
		return errors.New("ticker ID is required")
	}
	data, err := json.Marshal(ticker)
	if err != nil {
		return err
	}
	return rm.db.Put(ticker.ID, data)
}

func (rm *ReferenceManager) DeleteTicker(id string) error {
	return rm.db.Delete(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, id))
}

func (rm *ReferenceManager) GetTicker(id string) (*TickerReference, error) {
	var ticker TickerReference
	err := rm.db.Get(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, id), ticker)
	if err != nil {
		return nil, err
	}
	return &ticker, nil
}

// func (rm *ReferenceManager) GetAllTickers() ([]TickerReference, error) {
// 	iter := rm.db.NewIterator(nil, nil)
// 	defer iter.Release()

// 	var tickers []TickerReference
// 	for iter.Next() {
// 		var ticker TickerReference
// 		err := json.Unmarshal(iter.Value(), &ticker)
// 		if err != nil {
// 			return nil, err
// 		}
// 		tickers = append(tickers, ticker)
// 	}
// 	return tickers, iter.Error()
// }
