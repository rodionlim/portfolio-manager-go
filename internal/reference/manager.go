package reference

import (
	"errors"
	"fmt"
	"os"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"

	"gopkg.in/yaml.v2"
)

type ReferenceManager struct {
	db dal.Database
}

func NewReferenceManager(db dal.Database, filePath string) (*ReferenceManager, error) {
	rm := &ReferenceManager{db: db}

	// Check if the database is empty and seed it if necessary
	isEmpty, err := rm.isDatabaseEmpty()
	if err != nil {
		return nil, err
	}
	if isEmpty {
		err = rm.seedReferenceData(filePath)
		if err != nil {
			return nil, err
		}
	}

	return rm, nil
}

func (rm *ReferenceManager) isDatabaseEmpty() (bool, error) {
	refKeys, err := rm.db.GetAllKeysWithPrefix(string(types.ReferenceDataKeyPrefix))
	if err != nil {
		return false, err
	}
	return len(refKeys) == 0, nil
}

func (rm *ReferenceManager) seedReferenceData(filePath string) error {
	if filePath == "" {
		return nil // no seed file provided
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var tickers []TickerReference
	err = yaml.Unmarshal(data, &tickers)
	if err != nil {
		return err
	}

	for _, ticker := range tickers {
		_, err := rm.AddTicker(ticker)
		if err != nil {
			return err
		}
	}
	logging.GetLogger().Info("Seeded initial reference data from YAML file", filePath)
	return nil
}

func (rm *ReferenceManager) AddTicker(ticker TickerReference) (string, error) {
	err := rm.db.Put(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, ticker.ID), ticker)
	if err != nil {
		return "", err
	}
	return ticker.ID, nil
}

func (rm *ReferenceManager) UpdateTicker(ticker *TickerReference) error {
	if ticker.ID == "" {
		return errors.New("ticker ID is required")
	}
	return rm.db.Put(ticker.ID, ticker)
}

func (rm *ReferenceManager) DeleteTicker(id string) error {
	return rm.db.Delete(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, id))
}

func (rm *ReferenceManager) GetTicker(id string) (TickerReference, error) {
	var ticker TickerReference
	err := rm.db.Get(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, id), ticker)
	if err != nil {
		return TickerReference{}, err
	}
	return ticker, nil
}

func (rm *ReferenceManager) GetRefData() (map[string]TickerReference, error) {
	refKeys, err := rm.db.GetAllKeysWithPrefix(string(types.ReferenceDataKeyPrefix))
	if err != nil {
		return nil, err
	}

	refs := make(map[string]TickerReference)
	for _, key := range refKeys {
		var ref TickerReference
		err := rm.db.Get(key, &ref)
		if err != nil {
			return nil, err
		}
		refs[ref.ID] = ref
	}

	logging.GetLogger().Info("Loaded ticker references from database")

	return refs, nil
}
