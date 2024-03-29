// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/surahman/FTeX/pkg/quotes (interfaces: Quotes)

// Package quotes is a generated GoMock package.
package quotes

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	decimal "github.com/shopspring/decimal"
	models "github.com/surahman/FTeX/pkg/models"
)

// MockQuotes is a mock of Quotes interface.
type MockQuotes struct {
	ctrl     *gomock.Controller
	recorder *MockQuotesMockRecorder
}

// MockQuotesMockRecorder is the mock recorder for MockQuotes.
type MockQuotesMockRecorder struct {
	mock *MockQuotes
}

// NewMockQuotes creates a new mock instance.
func NewMockQuotes(ctrl *gomock.Controller) *MockQuotes {
	mock := &MockQuotes{ctrl: ctrl}
	mock.recorder = &MockQuotesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuotes) EXPECT() *MockQuotesMockRecorder {
	return m.recorder
}

// CryptoConversion mocks base method.
func (m *MockQuotes) CryptoConversion(arg0, arg1 string, arg2 decimal.Decimal, arg3 bool, arg4 func(string, string) (models.CryptoQuote, error)) (decimal.Decimal, decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CryptoConversion", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(decimal.Decimal)
	ret1, _ := ret[1].(decimal.Decimal)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CryptoConversion indicates an expected call of CryptoConversion.
func (mr *MockQuotesMockRecorder) CryptoConversion(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoConversion", reflect.TypeOf((*MockQuotes)(nil).CryptoConversion), arg0, arg1, arg2, arg3, arg4)
}

// FiatConversion mocks base method.
func (m *MockQuotes) FiatConversion(arg0, arg1 string, arg2 decimal.Decimal, arg3 func(string, string, decimal.Decimal) (models.FiatQuote, error)) (decimal.Decimal, decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatConversion", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(decimal.Decimal)
	ret1, _ := ret[1].(decimal.Decimal)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FiatConversion indicates an expected call of FiatConversion.
func (mr *MockQuotesMockRecorder) FiatConversion(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatConversion", reflect.TypeOf((*MockQuotes)(nil).FiatConversion), arg0, arg1, arg2, arg3)
}

// cryptoQuote mocks base method.
func (m *MockQuotes) cryptoQuote(arg0, arg1 string) (models.CryptoQuote, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "cryptoQuote", arg0, arg1)
	ret0, _ := ret[0].(models.CryptoQuote)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// cryptoQuote indicates an expected call of cryptoQuote.
func (mr *MockQuotesMockRecorder) cryptoQuote(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "cryptoQuote", reflect.TypeOf((*MockQuotes)(nil).cryptoQuote), arg0, arg1)
}

// fiatQuote mocks base method.
func (m *MockQuotes) fiatQuote(arg0, arg1 string, arg2 decimal.Decimal) (models.FiatQuote, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "fiatQuote", arg0, arg1, arg2)
	ret0, _ := ret[0].(models.FiatQuote)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// fiatQuote indicates an expected call of fiatQuote.
func (mr *MockQuotesMockRecorder) fiatQuote(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "fiatQuote", reflect.TypeOf((*MockQuotes)(nil).fiatQuote), arg0, arg1, arg2)
}
