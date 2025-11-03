package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// EventHandler handles file system events.
type EventHandler interface {
	OnCreated(filePath string)
}

// Watcher wraps fsnotify to provide file watching functionality.
type Watcher struct {
	watcher   *fsnotify.Watcher
	handler   EventHandler
	rootPath  string
	recursive bool
}

// NewWatcher creates a new file system watcher.
func NewWatcher(rootPath string, recursive bool, handler EventHandler) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	watcher := &Watcher{
		watcher:   w,
		handler:   handler,
		rootPath:  rootPath,
		recursive: recursive,
	}

	// Add root path and all subdirectories if recursive
	if recursive {
		if err := watcher.addRecursive(rootPath); err != nil {
			w.Close()
			return nil, fmt.Errorf("failed to add recursive paths: %w", err)
		}
	} else {
		if err := w.Add(rootPath); err != nil {
			w.Close()
			return nil, fmt.Errorf("failed to add root path: %w", err)
		}
	}

	return watcher, nil
}

// addRecursive adds a directory and all its subdirectories to the watcher.
func (w *Watcher) addRecursive(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return w.watcher.Add(p)
		}
		return nil
	})
}

// Start starts the watcher event loop.
func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				w.handleEvent(event)
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()
}

// handleEvent processes file system events.
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only handle Create events (matching Python's on_created)
	if event.Op&fsnotify.Create == fsnotify.Create {
		// Check if it's a directory - if so, add it to watcher if recursive
		info, err := os.Stat(event.Name)
		if err != nil {
			log.Printf("failed to stat %s: %v", event.Name, err)
			return
		}

		if info.IsDir() {
			// Add new directory to watcher if recursive
			if w.recursive {
				if err := w.addRecursive(event.Name); err != nil {
					log.Printf("failed to add new directory to watcher: %v", err)
				}
			}
			return
		}

		// It's a file, check if we should process it
		// Filter out .collective directories (matching Python logic)
		if !strings.Contains(event.Name, ".collective") {
			w.handler.OnCreated(event.Name)
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

