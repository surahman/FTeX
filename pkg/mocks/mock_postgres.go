// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/surahman/FTeX/pkg/postgres (interfaces: Postgres)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	uuid "github.com/gofrs/uuid"
	gomock "github.com/golang/mock/gomock"
	pgtype "github.com/jackc/pgx/v5/pgtype"
	decimal "github.com/shopspring/decimal"
	models "github.com/surahman/FTeX/pkg/models/postgres"
	postgres "github.com/surahman/FTeX/pkg/postgres"
)

// MockPostgres is a mock of Postgres interface.
type MockPostgres struct {
	ctrl     *gomock.Controller
	recorder *MockPostgresMockRecorder
}

// MockPostgresMockRecorder is the mock recorder for MockPostgres.
type MockPostgresMockRecorder struct {
	mock *MockPostgres
}

// NewMockPostgres creates a new mock instance.
func NewMockPostgres(ctrl *gomock.Controller) *MockPostgres {
	mock := &MockPostgres{ctrl: ctrl}
	mock.recorder = &MockPostgresMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPostgres) EXPECT() *MockPostgresMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockPostgres) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockPostgresMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPostgres)(nil).Close))
}

// CryptoBalanceCurrency mocks base method.
func (m *MockPostgres) CryptoBalanceCurrency(arg0 uuid.UUID, arg1 string) (postgres.CryptoAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CryptoBalanceCurrency", arg0, arg1)
	ret0, _ := ret[0].(postgres.CryptoAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CryptoBalanceCurrency indicates an expected call of CryptoBalanceCurrency.
func (mr *MockPostgresMockRecorder) CryptoBalanceCurrency(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoBalanceCurrency", reflect.TypeOf((*MockPostgres)(nil).CryptoBalanceCurrency), arg0, arg1)
}

// CryptoCreateAccount mocks base method.
func (m *MockPostgres) CryptoCreateAccount(arg0 uuid.UUID, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CryptoCreateAccount", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CryptoCreateAccount indicates an expected call of CryptoCreateAccount.
func (mr *MockPostgresMockRecorder) CryptoCreateAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoCreateAccount", reflect.TypeOf((*MockPostgres)(nil).CryptoCreateAccount), arg0, arg1)
}

// CryptoPurchase mocks base method.
func (m *MockPostgres) CryptoPurchase(arg0 uuid.UUID, arg1 postgres.Currency, arg2 decimal.Decimal, arg3 string, arg4 decimal.Decimal) (*postgres.FiatJournal, *postgres.CryptoJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CryptoPurchase", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*postgres.FiatJournal)
	ret1, _ := ret[1].(*postgres.CryptoJournal)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CryptoPurchase indicates an expected call of CryptoPurchase.
func (mr *MockPostgresMockRecorder) CryptoPurchase(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoPurchase", reflect.TypeOf((*MockPostgres)(nil).CryptoPurchase), arg0, arg1, arg2, arg3, arg4)
}

// CryptoTxDetailsCurrency mocks base method.
func (m *MockPostgres) CryptoTxDetailsCurrency(arg0, arg1 uuid.UUID) ([]postgres.CryptoJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CryptoTxDetailsCurrency", arg0, arg1)
	ret0, _ := ret[0].([]postgres.CryptoJournal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CryptoTxDetailsCurrency indicates an expected call of CryptoTxDetailsCurrency.
func (mr *MockPostgresMockRecorder) CryptoTxDetailsCurrency(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoTxDetailsCurrency", reflect.TypeOf((*MockPostgres)(nil).CryptoTxDetailsCurrency), arg0, arg1)
}

// FiatBalanceCurrency mocks base method.
func (m *MockPostgres) FiatBalanceCurrency(arg0 uuid.UUID, arg1 postgres.Currency) (postgres.FiatAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatBalanceCurrency", arg0, arg1)
	ret0, _ := ret[0].(postgres.FiatAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatBalanceCurrency indicates an expected call of FiatBalanceCurrency.
func (mr *MockPostgresMockRecorder) FiatBalanceCurrency(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatBalanceCurrency", reflect.TypeOf((*MockPostgres)(nil).FiatBalanceCurrency), arg0, arg1)
}

// FiatBalanceCurrencyPaginated mocks base method.
func (m *MockPostgres) FiatBalanceCurrencyPaginated(arg0 uuid.UUID, arg1 postgres.Currency, arg2 int32) ([]postgres.FiatAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatBalanceCurrencyPaginated", arg0, arg1, arg2)
	ret0, _ := ret[0].([]postgres.FiatAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatBalanceCurrencyPaginated indicates an expected call of FiatBalanceCurrencyPaginated.
func (mr *MockPostgresMockRecorder) FiatBalanceCurrencyPaginated(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatBalanceCurrencyPaginated", reflect.TypeOf((*MockPostgres)(nil).FiatBalanceCurrencyPaginated), arg0, arg1, arg2)
}

// FiatCreateAccount mocks base method.
func (m *MockPostgres) FiatCreateAccount(arg0 uuid.UUID, arg1 postgres.Currency) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatCreateAccount", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// FiatCreateAccount indicates an expected call of FiatCreateAccount.
func (mr *MockPostgresMockRecorder) FiatCreateAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatCreateAccount", reflect.TypeOf((*MockPostgres)(nil).FiatCreateAccount), arg0, arg1)
}

// FiatExternalTransfer mocks base method.
func (m *MockPostgres) FiatExternalTransfer(arg0 context.Context, arg1 *postgres.FiatTransactionDetails) (*postgres.FiatAccountTransferResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatExternalTransfer", arg0, arg1)
	ret0, _ := ret[0].(*postgres.FiatAccountTransferResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatExternalTransfer indicates an expected call of FiatExternalTransfer.
func (mr *MockPostgresMockRecorder) FiatExternalTransfer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatExternalTransfer", reflect.TypeOf((*MockPostgres)(nil).FiatExternalTransfer), arg0, arg1)
}

// FiatInternalTransfer mocks base method.
func (m *MockPostgres) FiatInternalTransfer(arg0 context.Context, arg1, arg2 *postgres.FiatTransactionDetails) (*postgres.FiatAccountTransferResult, *postgres.FiatAccountTransferResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatInternalTransfer", arg0, arg1, arg2)
	ret0, _ := ret[0].(*postgres.FiatAccountTransferResult)
	ret1, _ := ret[1].(*postgres.FiatAccountTransferResult)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FiatInternalTransfer indicates an expected call of FiatInternalTransfer.
func (mr *MockPostgresMockRecorder) FiatInternalTransfer(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatInternalTransfer", reflect.TypeOf((*MockPostgres)(nil).FiatInternalTransfer), arg0, arg1, arg2)
}

// FiatTransactionsCurrencyPaginated mocks base method.
func (m *MockPostgres) FiatTransactionsCurrencyPaginated(arg0 uuid.UUID, arg1 postgres.Currency, arg2, arg3 int32, arg4, arg5 pgtype.Timestamptz) ([]postgres.FiatJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatTransactionsCurrencyPaginated", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].([]postgres.FiatJournal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatTransactionsCurrencyPaginated indicates an expected call of FiatTransactionsCurrencyPaginated.
func (mr *MockPostgresMockRecorder) FiatTransactionsCurrencyPaginated(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatTransactionsCurrencyPaginated", reflect.TypeOf((*MockPostgres)(nil).FiatTransactionsCurrencyPaginated), arg0, arg1, arg2, arg3, arg4, arg5)
}

// FiatTxDetailsCurrency mocks base method.
func (m *MockPostgres) FiatTxDetailsCurrency(arg0, arg1 uuid.UUID) ([]postgres.FiatJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatTxDetailsCurrency", arg0, arg1)
	ret0, _ := ret[0].([]postgres.FiatJournal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatTxDetailsCurrency indicates an expected call of FiatTxDetailsCurrency.
func (mr *MockPostgresMockRecorder) FiatTxDetailsCurrency(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatTxDetailsCurrency", reflect.TypeOf((*MockPostgres)(nil).FiatTxDetailsCurrency), arg0, arg1)
}

// Healthcheck mocks base method.
func (m *MockPostgres) Healthcheck() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Healthcheck")
	ret0, _ := ret[0].(error)
	return ret0
}

// Healthcheck indicates an expected call of Healthcheck.
func (mr *MockPostgresMockRecorder) Healthcheck() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Healthcheck", reflect.TypeOf((*MockPostgres)(nil).Healthcheck))
}

// Open mocks base method.
func (m *MockPostgres) Open() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Open")
	ret0, _ := ret[0].(error)
	return ret0
}

// Open indicates an expected call of Open.
func (mr *MockPostgresMockRecorder) Open() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockPostgres)(nil).Open))
}

// UserCredentials mocks base method.
func (m *MockPostgres) UserCredentials(arg0 string) (uuid.UUID, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserCredentials", arg0)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// UserCredentials indicates an expected call of UserCredentials.
func (mr *MockPostgresMockRecorder) UserCredentials(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserCredentials", reflect.TypeOf((*MockPostgres)(nil).UserCredentials), arg0)
}

// UserDelete mocks base method.
func (m *MockPostgres) UserDelete(arg0 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserDelete", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UserDelete indicates an expected call of UserDelete.
func (mr *MockPostgresMockRecorder) UserDelete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserDelete", reflect.TypeOf((*MockPostgres)(nil).UserDelete), arg0)
}

// UserGetInfo mocks base method.
func (m *MockPostgres) UserGetInfo(arg0 uuid.UUID) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserGetInfo", arg0)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UserGetInfo indicates an expected call of UserGetInfo.
func (mr *MockPostgresMockRecorder) UserGetInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserGetInfo", reflect.TypeOf((*MockPostgres)(nil).UserGetInfo), arg0)
}

// UserRegister mocks base method.
func (m *MockPostgres) UserRegister(arg0 *models.UserAccount) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserRegister", arg0)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UserRegister indicates an expected call of UserRegister.
func (mr *MockPostgresMockRecorder) UserRegister(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserRegister", reflect.TypeOf((*MockPostgres)(nil).UserRegister), arg0)
}
