package auth

import "github.com/google/uuid"

type TokenGenerator interface {
	Generate() string
}

type UuidGen struct{}

func (u *UuidGen) Generate() string {
	return uuid.New().String()
}

type IdGenerator interface {
	Generate() []byte
}

type ByteGenerator struct{}

func (u *ByteGenerator) Generate() []byte {
	id := uuid.New()
	return id[:]
}
