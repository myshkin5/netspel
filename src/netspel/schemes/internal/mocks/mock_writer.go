package mocks

import "netspel/factory"

type MockWriter struct {
	Messages chan []byte
}

func NewMockWriter() *MockWriter {
	return &MockWriter{
		Messages: make(chan []byte, 10000),
	}
}

func (m *MockWriter) Init(config factory.Config) error {
	return nil
}

func (m *MockWriter) Write(message []byte) (int, error) {
	m.Messages <- message
	return len(message), nil
}