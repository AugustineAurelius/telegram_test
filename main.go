package main

import (
	"context"
	"flag"
	"sync"

	"go.uber.org/zap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
)

// memorySession implements in-memory session storage.
// Goroutine-safe.
type memorySession struct {
	mux  sync.RWMutex
	data []byte
}

// LoadSession loads session from memory.
func (s *memorySession) LoadSession(context.Context) ([]byte, error) {
	if s == nil {
		return nil, session.ErrNotFound
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	if len(s.data) == 0 {
		return nil, session.ErrNotFound
	}

	cpy := append([]byte(nil), s.data...)

	return cpy, nil
}

// StoreSession stores session to memory.
func (s *memorySession) StoreSession(ctx context.Context, data []byte) error {
	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
	return nil
}

func main() {
	// Grab those from https://my.telegram.org/apps.
	appID := flag.Int("api-id", 27138333, "app id")
	appHash := flag.String("api-hash", "7a06ac199afe8c1fea10f26d098b6870", "app hash")
	// Get it from bot father.
	token := flag.String("token", "6611877499:AAEr1O9_UcS0GVOomgt0y7yhuDGDENTJpsQ", "bot token")
	flag.Parse()

	// Using custom session storage.
	// You can save session to database, e.g. Redis, MongoDB or postgres.
	// See memorySession for implementation details.
	sessionStorage := &memorySession{}

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		client := telegram.NewClient(*appID, *appHash, telegram.Options{
			SessionStorage: sessionStorage,
			Logger:         log,
		})

		return client.Run(ctx, func(ctx context.Context) error {
			// Checking auth status.
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return err
			}
			// Can be already authenticated if we have valid session in
			// session storage.
			if !status.Authorized {
				// Otherwise, perform bot authentication.
				if _, err := client.Auth().Bot(ctx, *token); err != nil {
					return err
				}
			}

			// All good, manually authenticated.
			log.Info("Done")

			return nil
		})
	})
}
