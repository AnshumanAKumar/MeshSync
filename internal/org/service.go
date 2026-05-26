package org

import (
	"crypto/rand"
	"math/big"
	"time"

	"meshsync/internal/models"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) CreateOrg(name string, passcode string) (*models.Org, error) {
	if passcode == "" {
		passcode = generatePasscode()
	}

	org := &models.Org{
		Name:      name,
		Passcode:  passcode,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	return org, nil
}

func generatePasscode() string {

	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 6)

	for i := range b {

		randomIndex, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(charset))),
		)

		if err != nil {
			panic(err)
		}

		b[i] = charset[randomIndex.Int64()]
	}

	return string(b)
}
