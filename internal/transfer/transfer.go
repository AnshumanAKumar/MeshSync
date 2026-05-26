package transfer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"meshsync/internal/models"
)

type TransferService struct {
	baseDir string
	port    int
}

func NewTransferService(
	baseDir string,
	port int,
) *TransferService {

	return &TransferService{
		baseDir: baseDir,
		port:    port,
	}
}

func (ts *TransferService) StartServer(
	ctx context.Context,
	port int,
	events chan<- *models.TransferEvent,
) error {

	if _, err := os.Stat(ts.baseDir); os.IsNotExist(err) {

		if err := os.MkdirAll(
			ts.baseDir,
			0755,
		); err != nil {

			return fmt.Errorf(
				"failed creating transfer directory: %w",
				err,
			)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/transfer/download",
		func(w http.ResponseWriter, r *http.Request) {
			ts.handleDownload(w, r)
		},
	)

	mux.HandleFunc(
		"/api/v1/transfer/upload",
		func(w http.ResponseWriter, r *http.Request) {
			ts.handleUpload(w, r, events)
		},
	)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fmt.Printf(
		"[TRANSFER] server starting port=%d\n",
		port,
	)

	go func() {

		<-ctx.Done()

		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

func (ts *TransferService) sanitizePath(
	requestPath string,
) (string, error) {

	cleanPath := filepath.Clean(
		requestPath,
	)

	fullPath := filepath.Join(
		ts.baseDir,
		cleanPath,
	)

	absBase, err := filepath.Abs(
		ts.baseDir,
	)

	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(
		fullPath,
	)

	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(
		absPath,
		absBase,
	) {

		return "", fmt.Errorf(
			"invalid path",
		)
	}

	return absPath, nil
}

func (ts *TransferService) handleDownload(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {

		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	filePath := r.URL.Query().Get(
		"file",
	)

	if filePath == "" {

		http.Error(
			w,
			"file parameter required",
			http.StatusBadRequest,
		)

		return
	}

	fullPath, err := ts.sanitizePath(
		filePath,
	)

	if err != nil {

		http.Error(
			w,
			"invalid path",
			http.StatusForbidden,
		)

		return
	}

	fileInfo, err := os.Stat(
		fullPath,
	)

	if err != nil || fileInfo.IsDir() {

		http.Error(
			w,
			"file not found",
			http.StatusNotFound,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/octet-stream",
	)

	w.Header().Set(
		"Content-Length",
		fmt.Sprintf("%d", fileInfo.Size()),
	)

	http.ServeFile(
		w,
		r,
		fullPath,
	)
}

func (ts *TransferService) handleUpload(
	w http.ResponseWriter,
	r *http.Request,
	events chan<- *models.TransferEvent,
) {

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	filePath := r.URL.Query().Get(
		"file",
	)

	if filePath == "" {

		http.Error(
			w,
			"file parameter required",
			http.StatusBadRequest,
		)

		return
	}

	fullPath, err := ts.sanitizePath(
		filePath,
	)

	if err != nil {

		http.Error(
			w,
			"invalid path",
			http.StatusForbidden,
		)

		return
	}

	dir := filepath.Dir(
		fullPath,
	)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {

		http.Error(
			w,
			"failed creating directory",
			http.StatusInternalServerError,
		)

		return
	}

	file, err := os.Create(
		fullPath,
	)

	if err != nil {

		http.Error(
			w,
			"failed creating file",
			http.StatusInternalServerError,
		)

		return
	}

	defer file.Close()

	written, err := io.Copy(
		file,
		r.Body,
	)

	if err != nil {

		http.Error(
			w,
			"failed writing file",
			http.StatusInternalServerError,
		)

		return
	}

	peerID := r.Header.Get(
		"X-Peer-ID",
	)

	if peerID != "" {

		select {

		case events <- &models.TransferEvent{
			PeerID:    peerID,
			FilePath:  filePath,
			FileSize:  written,
			Status:    "completed",
			Timestamp: time.Now().UnixMilli(),
		}:

		default:
		}
	}

	w.WriteHeader(
		http.StatusOK,
	)
}
