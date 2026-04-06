package teable

import (
	"github.com/sennadevos/kosa/internal/backend"
	"github.com/sennadevos/kosa/internal/config"
)

func init() {
	backend.Register("teable", func(cfg *config.Config) (backend.Backend, error) {
		return New(cfg.Backend.Teable), nil
	})
}
