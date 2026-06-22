package mocks

import "portfolio-manager/pkg/types"

type MockMarketDataScreener struct{}

func (m *MockMarketDataScreener) FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchUSAIndustryStocksOverview(string) ([]types.USAIndustryStockOverview, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchUSAIndustryStocksPerformance(string) ([]types.USAIndustryStockPerformance, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFLargestInflowsOverview() ([]types.ETFFundFlowOverview, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFLargestInflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFLargestInflowsFundFlows() ([]types.ETFFundFlows, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFLargestOutflowsOverview() ([]types.ETFFundFlowOverview, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFLargestOutflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFLargestOutflowsFundFlows() ([]types.ETFFundFlows, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFSectorOverview() ([]types.ETFSectorOverview, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFSectorPerformance() ([]types.ETFFundFlowPerformance, error) {
	return nil, nil
}

func (m *MockMarketDataScreener) FetchETFSectorFundFlows() ([]types.ETFFundFlows, error) {
	return nil, nil
}
