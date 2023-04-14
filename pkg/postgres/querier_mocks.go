// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/surahman/FTeX/pkg/postgres (interfaces: Querier)

// Package postgres is a generated GoMock package.
package postgres

import (
	context "context"
	reflect "reflect"

	uuid "github.com/gofrs/uuid"
	gomock "github.com/golang/mock/gomock"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	decimal "github.com/shopspring/decimal"
)

// MockQuerier is a mock of Querier interface.
type MockQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockQuerierMockRecorder
}

// MockQuerierMockRecorder is the mock recorder for MockQuerier.
type MockQuerierMockRecorder struct {
	mock *MockQuerier
}

// NewMockQuerier creates a new mock instance.
func NewMockQuerier(ctrl *gomock.Controller) *MockQuerier {
	mock := &MockQuerier{ctrl: ctrl}
	mock.recorder = &MockQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuerier) EXPECT() *MockQuerierMockRecorder {
	return m.recorder
}

// FiatCreateAccount mocks base method.
func (m *MockQuerier) FiatCreateAccount(arg0 context.Context, arg1 *FiatCreateAccountParams) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatCreateAccount", arg0, arg1)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatCreateAccount indicates an expected call of FiatCreateAccount.
func (mr *MockQuerierMockRecorder) FiatCreateAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatCreateAccount", reflect.TypeOf((*MockQuerier)(nil).FiatCreateAccount), arg0, arg1)
}

// FiatExternalTransferJournalEntry mocks base method.
func (m *MockQuerier) FiatExternalTransferJournalEntry(arg0 context.Context, arg1 *FiatExternalTransferJournalEntryParams) (FiatExternalTransferJournalEntryRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatExternalTransferJournalEntry", arg0, arg1)
	ret0, _ := ret[0].(FiatExternalTransferJournalEntryRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatExternalTransferJournalEntry indicates an expected call of FiatExternalTransferJournalEntry.
func (mr *MockQuerierMockRecorder) FiatExternalTransferJournalEntry(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatExternalTransferJournalEntry", reflect.TypeOf((*MockQuerier)(nil).FiatExternalTransferJournalEntry), arg0, arg1)
}

// FiatGetAccount mocks base method.
func (m *MockQuerier) FiatGetAccount(arg0 context.Context, arg1 *FiatGetAccountParams) (FiatAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatGetAccount", arg0, arg1)
	ret0, _ := ret[0].(FiatAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatGetAccount indicates an expected call of FiatGetAccount.
func (mr *MockQuerierMockRecorder) FiatGetAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatGetAccount", reflect.TypeOf((*MockQuerier)(nil).FiatGetAccount), arg0, arg1)
}

// FiatGetAllAccounts mocks base method.
func (m *MockQuerier) FiatGetAllAccounts(arg0 context.Context, arg1 uuid.UUID) ([]FiatAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatGetAllAccounts", arg0, arg1)
	ret0, _ := ret[0].([]FiatAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatGetAllAccounts indicates an expected call of FiatGetAllAccounts.
func (mr *MockQuerierMockRecorder) FiatGetAllAccounts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatGetAllAccounts", reflect.TypeOf((*MockQuerier)(nil).FiatGetAllAccounts), arg0, arg1)
}

// FiatGetJournalTransaction mocks base method.
func (m *MockQuerier) FiatGetJournalTransaction(arg0 context.Context, arg1 uuid.UUID) ([]FiatJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatGetJournalTransaction", arg0, arg1)
	ret0, _ := ret[0].([]FiatJournal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatGetJournalTransaction indicates an expected call of FiatGetJournalTransaction.
func (mr *MockQuerierMockRecorder) FiatGetJournalTransaction(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatGetJournalTransaction", reflect.TypeOf((*MockQuerier)(nil).FiatGetJournalTransaction), arg0, arg1)
}

// FiatGetJournalTransactionForAccount mocks base method.
func (m *MockQuerier) FiatGetJournalTransactionForAccount(arg0 context.Context, arg1 *FiatGetJournalTransactionForAccountParams) ([]FiatJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatGetJournalTransactionForAccount", arg0, arg1)
	ret0, _ := ret[0].([]FiatJournal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatGetJournalTransactionForAccount indicates an expected call of FiatGetJournalTransactionForAccount.
func (mr *MockQuerierMockRecorder) FiatGetJournalTransactionForAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatGetJournalTransactionForAccount", reflect.TypeOf((*MockQuerier)(nil).FiatGetJournalTransactionForAccount), arg0, arg1)
}

// FiatGetJournalTransactionForAccountBetweenDates mocks base method.
func (m *MockQuerier) FiatGetJournalTransactionForAccountBetweenDates(arg0 context.Context, arg1 *FiatGetJournalTransactionForAccountBetweenDatesParams) ([]FiatJournal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatGetJournalTransactionForAccountBetweenDates", arg0, arg1)
	ret0, _ := ret[0].([]FiatJournal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatGetJournalTransactionForAccountBetweenDates indicates an expected call of FiatGetJournalTransactionForAccountBetweenDates.
func (mr *MockQuerierMockRecorder) FiatGetJournalTransactionForAccountBetweenDates(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatGetJournalTransactionForAccountBetweenDates", reflect.TypeOf((*MockQuerier)(nil).FiatGetJournalTransactionForAccountBetweenDates), arg0, arg1)
}

// FiatInternalTransferJournalEntry mocks base method.
func (m *MockQuerier) FiatInternalTransferJournalEntry(arg0 context.Context, arg1 *FiatInternalTransferJournalEntryParams) (FiatInternalTransferJournalEntryRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatInternalTransferJournalEntry", arg0, arg1)
	ret0, _ := ret[0].(FiatInternalTransferJournalEntryRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatInternalTransferJournalEntry indicates an expected call of FiatInternalTransferJournalEntry.
func (mr *MockQuerierMockRecorder) FiatInternalTransferJournalEntry(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatInternalTransferJournalEntry", reflect.TypeOf((*MockQuerier)(nil).FiatInternalTransferJournalEntry), arg0, arg1)
}

// FiatRowLockAccount mocks base method.
func (m *MockQuerier) FiatRowLockAccount(arg0 context.Context, arg1 *FiatRowLockAccountParams) (decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatRowLockAccount", arg0, arg1)
	ret0, _ := ret[0].(decimal.Decimal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatRowLockAccount indicates an expected call of FiatRowLockAccount.
func (mr *MockQuerierMockRecorder) FiatRowLockAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatRowLockAccount", reflect.TypeOf((*MockQuerier)(nil).FiatRowLockAccount), arg0, arg1)
}

// FiatUpdateAccountBalance mocks base method.
func (m *MockQuerier) FiatUpdateAccountBalance(arg0 context.Context, arg1 *FiatUpdateAccountBalanceParams) (FiatUpdateAccountBalanceRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FiatUpdateAccountBalance", arg0, arg1)
	ret0, _ := ret[0].(FiatUpdateAccountBalanceRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FiatUpdateAccountBalance indicates an expected call of FiatUpdateAccountBalance.
func (mr *MockQuerierMockRecorder) FiatUpdateAccountBalance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FiatUpdateAccountBalance", reflect.TypeOf((*MockQuerier)(nil).FiatUpdateAccountBalance), arg0, arg1)
}

// testRoundHalfEven mocks base method.
func (m *MockQuerier) testRoundHalfEven(arg0 context.Context, arg1 *testRoundHalfEvenParams) (decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "testRoundHalfEven", arg0, arg1)
	ret0, _ := ret[0].(decimal.Decimal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// testRoundHalfEven indicates an expected call of testRoundHalfEven.
func (mr *MockQuerierMockRecorder) testRoundHalfEven(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "testRoundHalfEven", reflect.TypeOf((*MockQuerier)(nil).testRoundHalfEven), arg0, arg1)
}

// userCreate mocks base method.
func (m *MockQuerier) userCreate(arg0 context.Context, arg1 *userCreateParams) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "userCreate", arg0, arg1)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// userCreate indicates an expected call of userCreate.
func (mr *MockQuerierMockRecorder) userCreate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "userCreate", reflect.TypeOf((*MockQuerier)(nil).userCreate), arg0, arg1)
}

// userDelete mocks base method.
func (m *MockQuerier) userDelete(arg0 context.Context, arg1 string) (pgconn.CommandTag, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "userDelete", arg0, arg1)
	ret0, _ := ret[0].(pgconn.CommandTag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// userDelete indicates an expected call of userDelete.
func (mr *MockQuerierMockRecorder) userDelete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "userDelete", reflect.TypeOf((*MockQuerier)(nil).userDelete), arg0, arg1)
}

// userGetClientId mocks base method.
func (m *MockQuerier) userGetClientId(arg0 context.Context, arg1 string) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "userGetClientId", arg0, arg1)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// userGetClientId indicates an expected call of userGetClientId.
func (mr *MockQuerierMockRecorder) userGetClientId(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "userGetClientId", reflect.TypeOf((*MockQuerier)(nil).userGetClientId), arg0, arg1)
}

// userGetCredentials mocks base method.
func (m *MockQuerier) userGetCredentials(arg0 context.Context, arg1 string) (userGetCredentialsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "userGetCredentials", arg0, arg1)
	ret0, _ := ret[0].(userGetCredentialsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// userGetCredentials indicates an expected call of userGetCredentials.
func (mr *MockQuerierMockRecorder) userGetCredentials(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "userGetCredentials", reflect.TypeOf((*MockQuerier)(nil).userGetCredentials), arg0, arg1)
}

// userGetInfo mocks base method.
func (m *MockQuerier) userGetInfo(arg0 context.Context, arg1 string) (userGetInfoRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "userGetInfo", arg0, arg1)
	ret0, _ := ret[0].(userGetInfoRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// userGetInfo indicates an expected call of userGetInfo.
func (mr *MockQuerierMockRecorder) userGetInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "userGetInfo", reflect.TypeOf((*MockQuerier)(nil).userGetInfo), arg0, arg1)
}
