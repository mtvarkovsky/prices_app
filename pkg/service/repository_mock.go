// Code generated by MockGen. DO NOT EDIT.
// Source: prices.go

// Package service is a generated GoMock package.
package service

import (
	context "context"
	models "prices/pkg/models"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// CreateMany mocks base method.
func (m *MockRepository) CreateMany(ctx context.Context, prices []*models.Price) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMany", ctx, prices)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateMany indicates an expected call of CreateMany.
func (mr *MockRepositoryMockRecorder) CreateMany(ctx, prices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMany", reflect.TypeOf((*MockRepository)(nil).CreateMany), ctx, prices)
}

// Get mocks base method.
func (m *MockRepository) Get(ctx context.Context, id string) (*models.Price, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(*models.Price)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockRepositoryMockRecorder) Get(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRepository)(nil).Get), ctx, id)
}

// ImportFile mocks base method.
func (m *MockRepository) ImportFile(ctx context.Context, filePath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ImportFile", ctx, filePath)
	ret0, _ := ret[0].(error)
	return ret0
}

// ImportFile indicates an expected call of ImportFile.
func (mr *MockRepositoryMockRecorder) ImportFile(ctx, filePath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImportFile", reflect.TypeOf((*MockRepository)(nil).ImportFile), ctx, filePath)
}
