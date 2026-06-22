package testify

import (
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/mock"
)

type MockMarketDataScreener struct {
	mock.Mock
}

func (m *MockMarketDataScreener) FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.USAIndustryPerformance), args.Error(1)
}

func (m *MockMarketDataScreener) FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.USAIndustryOverview), args.Error(1)
}

func (m *MockMarketDataScreener) FetchUSAIndustryStocksOverview(industry string) ([]types.USAIndustryStockOverview, error) {
	args := m.Called(industry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.USAIndustryStockOverview), args.Error(1)
}

func (m *MockMarketDataScreener) FetchUSAIndustryStocksPerformance(industry string) ([]types.USAIndustryStockPerformance, error) {
	args := m.Called(industry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.USAIndustryStockPerformance), args.Error(1)
}

func (m *MockMarketDataScreener) FetchETFLargestInflowsOverview() ([]types.ETFFundFlowOverview, error) {
	return mockETFFundFlowOverview(m.Called())
}

func (m *MockMarketDataScreener) FetchETFLargestInflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	return mockETFFundFlowPerformance(m.Called())
}

func (m *MockMarketDataScreener) FetchETFLargestInflowsFundFlows() ([]types.ETFFundFlows, error) {
	return mockETFFundFlows(m.Called())
}

func (m *MockMarketDataScreener) FetchETFLargestOutflowsOverview() ([]types.ETFFundFlowOverview, error) {
	return mockETFFundFlowOverview(m.Called())
}

func (m *MockMarketDataScreener) FetchETFLargestOutflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	return mockETFFundFlowPerformance(m.Called())
}

func (m *MockMarketDataScreener) FetchETFLargestOutflowsFundFlows() ([]types.ETFFundFlows, error) {
	return mockETFFundFlows(m.Called())
}

func (m *MockMarketDataScreener) FetchETFSectorOverview() ([]types.ETFSectorOverview, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.ETFSectorOverview), args.Error(1)
}

func (m *MockMarketDataScreener) FetchETFSectorPerformance() ([]types.ETFFundFlowPerformance, error) {
	return mockETFFundFlowPerformance(m.Called())
}

func (m *MockMarketDataScreener) FetchETFSectorFundFlows() ([]types.ETFFundFlows, error) {
	return mockETFFundFlows(m.Called())
}

func mockETFFundFlowOverview(args mock.Arguments) ([]types.ETFFundFlowOverview, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.ETFFundFlowOverview), args.Error(1)
}

func mockETFFundFlowPerformance(args mock.Arguments) ([]types.ETFFundFlowPerformance, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.ETFFundFlowPerformance), args.Error(1)
}

func mockETFFundFlows(args mock.Arguments) ([]types.ETFFundFlows, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.ETFFundFlows), args.Error(1)
}
