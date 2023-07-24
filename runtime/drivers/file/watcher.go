package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

const batchInterval = 250 * time.Millisecond

// watcher implements a recursive, batching file watcher on top of fsnotify.
type watcher struct {
	root        string
	watcher     *fsnotify.Watcher
	done        chan struct{}
	err         error
	mu          sync.Mutex
	subscribers map[string]drivers.WatchCallback
	buffer      []drivers.WatchEvent
}

func newWatcher(root string) (*watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &watcher{
		root:        root,
		watcher:     fsw,
		done:        make(chan struct{}),
		subscribers: make(map[string]drivers.WatchCallback),
	}

	err = w.addDir(root, false)
	if err != nil {
		w.watcher.Close()
		return nil, err
	}

	go w.run()

	return w, nil
}

func (w *watcher) close() {
	w.closeWithErr(nil)
}

func (w *watcher) closeWithErr(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		// Already closed
		return
	default:
	}

	w.err = err

	err = w.watcher.Close()
	if w.err == nil {
		w.err = err
	}
	if w.err == nil {
		w.err = fmt.Errorf("file watcher closed")
	}
	close(w.done)
}

func (w *watcher) subscribe(ctx context.Context, fn drivers.WatchCallback) error {
	if w.err != nil {
		return w.err
	}

	id := fmt.Sprintf("%v", fn)
	w.mu.Lock()
	w.subscribers[id] = fn
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.subscribers, id)
		w.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return w.err
	}
}

func (w *watcher) flush() {
	if len(w.buffer) == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, fn := range w.subscribers {
		fn(w.buffer)
	}

	w.buffer = nil
}

func (w *watcher) run() {
	err := w.runInner()
	w.closeWithErr(err)
}

func (w *watcher) runInner() error {
	timer := time.NewTimer(batchInterval)
	timerActive := true
	for {
		select {
		case <-timer.C:
			timerActive = false
			w.flush()
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			return err
		case e, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			we := drivers.WatchEvent{}
			if e.Has(fsnotify.Create) || e.Has(fsnotify.Write) {
				we.Type = runtimev1.FileEvent_FILE_EVENT_WRITE
			} else if e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
				we.Type = runtimev1.FileEvent_FILE_EVENT_DELETE
			} else {
				continue
			}

			path, err := filepath.Rel(w.root, e.Name)
			if err != nil {
				return err
			}
			path = filepath.Join("/", path)
			we.Path = path

			if e.Has(fsnotify.Create) {
				info, err := os.Stat(e.Name)
				we.Dir = err == nil && info.IsDir()
			}

			w.buffer = append(w.buffer, we)

			// Calling addDir after appending to w.buffer, to sequence events correctly
			if we.Dir && e.Has(fsnotify.Create) {
				err = w.addDir(e.Name, true)
				if err != nil {
					return err
				}
			}

			// NOTE: See docs for timer.Reset() for context on why we need to check if the timer is active
			if timerActive && !timer.Stop() {
				<-timer.C
			}
			timer.Reset(batchInterval)
			timerActive = true
		}
	}
}

func (w *watcher) addDir(path string, replay bool) error {
	err := w.watcher.Add(path)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if replay {
			ep, err := filepath.Rel(w.root, filepath.Join(path, e.Name()))
			if err != nil {
				return err
			}
			ep = filepath.Join("/", ep)

			w.buffer = append(w.buffer, drivers.WatchEvent{
				Path: ep,
				Type: runtimev1.FileEvent_FILE_EVENT_WRITE,
				Dir:  e.IsDir(),
			})
		}

		if e.IsDir() {
			err := w.addDir(filepath.Join(path, e.Name()), replay)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
