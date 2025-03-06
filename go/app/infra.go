package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID        int    `db:"id" json:"-"`
	Name      string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	return &itemRepository{fileName: "items.json"}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-1: add an implementation to store an item
	f, err := os.Create(i.fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(item)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	return nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(filePath string, image []byte) error {
	err := os.WriteFile(filePath, image, 0644)
	if err != nil {
		return fmt.Errorf("failed to store image: %w", err)
	}

	return nil
}
