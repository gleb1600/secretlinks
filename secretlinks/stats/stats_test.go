package main

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStatsStorage struct {
	mock.Mock
}

func TestAddNewItem(t *testing.T) {
	statsStorage := NewStatsStorage()
	statsStorage.AddNewItem("newkey", time.Now())

	assert.Equal(t, 1, len(statsStorage.items))
	assert.Equal(t, "newkey", statsStorage.items["newkey"].LinkKey)
	assert.Equal(t, 0, len(statsStorage.items["newkey"].VisitTime))
	assert.WithinDuration(t, time.Now(), statsStorage.items["newkey"].CreateTime, time.Millisecond)
}

func TestAppendVisitTimeSuccess(t *testing.T) {
	statsStorage := NewStatsStorage()
	statsStorage.AddNewItem("newkey", time.Now())
	statsStorage.AppendVisitTime("newkey", time.Now())

	assert.Equal(t, 1, len(statsStorage.items))
	assert.Equal(t, "newkey", statsStorage.items["newkey"].LinkKey)
	assert.Equal(t, 1, len(statsStorage.items["newkey"].VisitTime))
}

func TestAppendVisitTimeWrongKey(t *testing.T) {
	statsStorage := NewStatsStorage()
	statsStorage.AddNewItem("newkey", time.Now())
	statsStorage.AppendVisitTime("newkey2", time.Now())

	assert.Equal(t, 2, len(statsStorage.items))
	assert.Equal(t, 0, len(statsStorage.items["newkey"].VisitTime))
}

// Mock KafkaReader
type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) AddNewItem(key string, t time.Time) {
	m.Called(key, t)
}

func (m *MockStorage) AppendVisitTime(key string, t time.Time) {
	m.Called(key, t)
}

func (m *MockStorage) ShowItem(key string) {
	m.Called(key)
}
