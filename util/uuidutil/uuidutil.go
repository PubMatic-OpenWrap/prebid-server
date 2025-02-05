package uuidutil

import (
	"github.com/gofrs/uuid"
)

type UUIDGenerator interface {
	Generate() (string, error)
}

type UUIDRandomGenerator struct{}

func (UUIDRandomGenerator) Generate() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// fakeUUIDGenerator implements the UUIDGenerator interface
type fakeUUIDGenerator struct {
	id  string
	err error
}

func NewFakeUUIDGenerator(id string, err error) UUIDGenerator {
	return &fakeUUIDGenerator{id: id, err: err}
}

func (f fakeUUIDGenerator) Generate() (string, error) {
	return f.id, f.err
}
