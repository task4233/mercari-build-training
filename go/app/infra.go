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

type Item struct {
	ID        int    `db:"id" json:"-"`
	Name      string `db:"name" json:"name"`
	Category  string `db:"category" json:"category"`
	ImageName string `db:"image_name" json:"image_name"`
}

type Items struct {
	Items []*Item `db:"items" json:"items"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) (*Items, error)
	Get(ctx context.Context, id int) (*Item, error)
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
	// NOTE: I don't think we need to open the file twice.
	items, err := i.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all items: %w", err)
	}
	items.Items = append(items.Items, item)

	// STEP 4-1: add an implementation to store an item
	f, err := os.Create(i.fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(items)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	return nil
}

func (i *itemRepository) GetAll(ctx context.Context) (*Items, error) {
	f, err := os.OpenFile(i.fileName, os.O_RDONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Items{Items: []*Item{}}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var items Items
	err = json.NewDecoder(f).Decode(&items)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return &items, nil
}

func (i *itemRepository) Get(ctx context.Context, id int) (*Item, error) {
	items, err := i.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all items: %w", err)
	}

	if len(items.Items) <= id {
		return nil, fmt.Errorf("item not found")
	}

	return items.Items[id], nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(filePath string, image []byte) error {
	err := os.WriteFile(filePath, image, 0644)
	if err != nil {
		return fmt.Errorf("failed to store image2: %w", err)
	}

	return nil
}
