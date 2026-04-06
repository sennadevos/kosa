package backend

import (
	"fmt"

	"github.com/sennadevos/kosa/internal/config"
)

// New creates a Backend based on the config's backend type.
// The returned interface is intentionally opaque — callers must not
// type-assert to a specific backend implementation.
func New(cfg *config.Config) (Backend, error) {
	switch cfg.Backend.Type {
	case "teable":
		// Import cycle prevention: the teable package imports backend,
		// so we can't import teable here. Instead, registration is done
		// via the Register function below.
		factory, ok := registry[cfg.Backend.Type]
		if !ok {
			return nil, fmt.Errorf("backend %q registered but factory not found", cfg.Backend.Type)
		}
		return factory(cfg)
	default:
		factory, ok := registry[cfg.Backend.Type]
		if !ok {
			return nil, fmt.Errorf("unknown backend type: %q", cfg.Backend.Type)
		}
		return factory(cfg)
	}
}

// Factory creates a Backend from config.
type Factory func(cfg *config.Config) (Backend, error)

var registry = map[string]Factory{}

// Register adds a backend factory. Called from init() in backend packages.
func Register(name string, f Factory) {
	registry[name] = f
}
