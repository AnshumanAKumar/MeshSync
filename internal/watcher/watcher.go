package watcher

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"meshsync/internal/models"

	"github.com/fsnotify/fsnotify"
)

type WatcherService struct {
	watchPath string
}

func NewWatcherService(watchPath string) *WatcherService {
	return &WatcherService{
		watchPath: watchPath,
	}
}

// StartWatcher monitors the filesystem for mutations and emits events.
func (w *WatcherService) StartWatcher(
	ctx context.Context,
	watchPath string,
	events chan<- *models.FileEvent,
) error {

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Ensure watch path exists
	if _, err := os.Stat(watchPath); os.IsNotExist(err) {
		if err := os.MkdirAll(watchPath, 0755); err != nil {
			return fmt.Errorf("failed to create watch directory: %w", err)
		}
	}

	// Add path to watcher
	if err := watcher.Add(watchPath); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	fmt.Printf("[WATCHER] started watching directory: %s\n", watchPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Skip directories
			fileInfo, err := os.Stat(event.Name)
			if err != nil || fileInfo.IsDir() {
				continue
			}

			// Determine event type
			var eventType string
			if event.Op&fsnotify.Create == fsnotify.Create {
				eventType = "CREATE"
			} else if event.Op&fsnotify.Write == fsnotify.Write {
				eventType = "MODIFY"
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				eventType = "DELETE"
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				eventType = "RENAME"
			} else {
				continue
			}

			// Get relative path
			relPath, err := filepath.Rel(watchPath, event.Name)
			if err != nil {
				continue
			}

			// Get file size
			fileSize := int64(0)
			if eventType != "DELETE" && eventType != "RENAME" {
				if stat, err := os.Stat(event.Name); err == nil {
					fileSize = stat.Size()
				}
			}

			// Create file event
			fileEvent := &models.FileEvent{
				EventType: eventType,
				FilePath:  relPath,
				FileSize:  fileSize,
				Timestamp: time.Now().UnixMilli(),
			}

			fmt.Printf(
				"[WATCHER] detected event type=%s path=%s size=%d\n",
				eventType,
				relPath,
				fileSize,
			)

			// Send event to channel
			select {
			case events <- fileEvent:
			case <-ctx.Done():
				return nil
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("[WATCHER] error: %v\n", err)

		case <-ctx.Done():
			fmt.Println("[WATCHER] shutting down")
			return nil
		}
	}
}

// CalculateChecksum computes MD5 hash of a file
func (w *WatcherService) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
