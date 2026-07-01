package db

import (
	"github.com/puppe1990/cais/internal/models"
	"github.com/puppe1990/cais/internal/store"
)

// RunSeeds populates demo data. Safe to run multiple times.
func RunSeeds(s store.Store) error {
	if _, err := s.InsertContact(models.Contact{
		Name:  "Demo",
		Email: "demo@example.com",
	}); err != nil {
		return err
	}
	return nil
}
