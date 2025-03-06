package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	// Port is the port number to listen on.
	Port string
	// ImageDirPath is the path to the directory storing images.
	ImageDirPath string
}

// Run is a method to start the server.
// This method returns 0 if the server started successfully, and 1 otherwise.
func (s Server) Run() int {
	// set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	// STEP 4-6: set the log level to DEBUG
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// STEP 5-1: set up the database connection

	// set up handlers
	itemRepo := NewItemRepository()
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Hello)
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /items/{id}", h.GetItem)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)

	// start the server
	slog.Info("http server started on", "port", s.Port)
	err := http.ListenAndServe(":"+s.Port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
		return 1
	}

	return 0
}

type Handlers struct {
	// imgDirPath is the path to the directory storing images.
	imgDirPath string
	itemRepo   ItemRepository
}

type HelloResponse struct {
	Message string `json:"message"`
}

// Hello is a handler to return a Hello, world! message for GET / .
func (s *Handlers) Hello(w http.ResponseWriter, r *http.Request) {
	resp := HelloResponse{Message: "Hello, world!"}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type AddItemRequest struct {
	Name     string `form:"name"`
	Category string `form:"category"`
	Image    []byte `form:"image"`
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	name, category := r.FormValue("name"), r.FormValue("category")
	f, _, err := r.FormFile("image")
	if err != nil {
		return nil, fmt.Errorf("failed to get image value: %w", err)
	}
	defer f.Close()

	image, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file info: %w", err)
	}

	// validate the request
	if name == "" {
		return nil, errors.New("name is required")
	}
	if category == "" {
		return nil, errors.New("category is required")
	}
	if len(image) == 0 {
		return nil, errors.New("image is required")
	}

	return &AddItemRequest{
		Name:     name,
		Category: category,
		// Image:    image,
	}, nil
}

type GetItemsResponse struct {
	Items []*Item `json:"items"`
}

// GetItems is a handler to fetch all items for GET /items.
func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	items, err := s.itemRepo.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get items", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		slog.Error("failed to encode", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AddItem is a handler to add a new item for POST /items .
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileName, err := s.storeImage(req.Image)
	if err != nil {
		slog.Error("failed to store image: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item := &Item{
		Name:      req.Name,
		Category:  req.Category,
		ImageName: fileName,
	}
	message := fmt.Sprintf("item received: name: %s, category: %s, image_name: %s", item.Name, item.Category, item.ImageName)
	slog.Info(message)

	err = s.itemRepo.Insert(ctx, item)
	if err != nil {
		slog.Error("failed to store item: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := AddItemResponse{Message: message}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	hash := sha256.Sum256(image)
	filePath, err = s.buildImagePath(hex.EncodeToString(hash[:]) + ".jpg")
	if err != nil {
		return "", err
	}

	_, err = os.Stat(filePath)
	// as the file already exists, we can simply return the file path.
	if err == nil {
		return filePath, nil
	}
	// if the error is not not exist error, it should be unexpected error.
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to build image path with unexpected error: %w", err)
	}

	err = StoreImage(filePath, image)
	if err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}

	return filePath, nil
}

type GetItemRequest struct {
	ID int // converted from path value
}

// parseGetItemRequest parses and validates the request to get an item.
func parseGetItemRequest(r *http.Request) (*GetItemRequest, error) {
	id := r.PathValue("id")

	// validate the request
	n, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("id should be a number: %w", err)
	}
	if n < 0 {
		return nil, errors.New("id should be a positive number")
	}

	req := &GetItemRequest{
		ID: n, // from path parameter
	}

	return req, nil
}

type GetItemResponse struct {
	Item *Item `json:"item"`
}

func (s *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseGetItemRequest(r)
	if err != nil {
		slog.Warn("failed to parse get item request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := s.itemRepo.Get(ctx, req.ID)
	if err != nil {
		slog.Error("failed to get item", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		slog.Error("failed to encode", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type GetImageRequest struct {
	FileName string // path value
}

// parseGetImageRequest parses and validates the request to get an image.
func parseGetImageRequest(r *http.Request) (*GetImageRequest, error) {
	req := &GetImageRequest{
		FileName: r.PathValue("filename"), // from path parameter
	}

	// validate the request
	if req.FileName == "" {
		return nil, errors.New("filename is required")
	}

	return req, nil
}

// GetImage is a handler to return an image for GET /images/{filename} .
// If the specified image is not found, it returns the default image.
func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetImageRequest(r)
	if err != nil {
		slog.Warn("failed to parse get image request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imgPath, err := s.buildImagePath(req.FileName)
	if err != nil {
		slog.Warn("failed to build image path: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if the image exists
	_, err = os.Stat(imgPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Warn("failed to build image path: ", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// when the image is not found, it returns the default image without an error.
		slog.Debug("image not found", "filename", imgPath)
		imgPath = filepath.Join(s.imgDirPath, "default.jpg")
	}

	slog.Info("returned image", "path", imgPath)
	http.ServeFile(w, r, imgPath)
}

// buildImagePath builds the image path and validates it.
func (s *Handlers) buildImagePath(imageFileName string) (string, error) {
	imgPath := filepath.Join(s.imgDirPath, filepath.Clean(imageFileName))

	// to prevent directory traversal attacks
	rel, err := filepath.Rel(s.imgDirPath, imgPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid image path: %s", imgPath)
	}

	// validate the image suffix
	if !strings.HasSuffix(imgPath, ".jpg") && !strings.HasSuffix(imgPath, ".jpeg") {
		return "", fmt.Errorf("image path does not end with .jpg or .jpeg: %s", imgPath)
	}

	return imgPath, nil
}
