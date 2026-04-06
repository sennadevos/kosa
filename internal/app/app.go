package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/sennadevos/kosa/internal/backend"
	"github.com/sennadevos/kosa/internal/config"
	"github.com/sennadevos/kosa/internal/domain"
)

// App holds the backend and config, and provides all application-level operations.
type App struct {
	Backend backend.Backend
	Config  *config.Config
}

func New(b backend.Backend, cfg *config.Config) *App {
	return &App{Backend: b, Config: cfg}
}

// ResolveAccountID resolves an account name to an ID.
// If name is empty, uses the default from config.
func (a *App) ResolveAccountID(ctx context.Context, name string) (string, error) {
	if name == "" {
		name = a.Config.Defaults.Account
	}
	if name == "" {
		return "", fmt.Errorf("no account specified and no default configured")
	}
	accounts, err := a.Backend.ListAccounts(ctx, domain.AccountFilter{})
	if err != nil {
		return "", fmt.Errorf("listing accounts: %w", err)
	}
	for _, acc := range accounts {
		if strings.EqualFold(acc.Name, name) {
			return acc.ID, nil
		}
	}
	return "", fmt.Errorf("account %q not found", name)
}

// resolveCategoryID resolves a category name to an ID. Returns empty if name is empty.
func (a *App) resolveCategoryID(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	cat, err := a.Backend.GetCategoryByName(ctx, name)
	if err != nil {
		return "", err
	}
	return cat.ID, nil
}

// resolveTagIDs resolves tag names to IDs, creating tags that don't exist.
func (a *App) resolveTagIDs(ctx context.Context, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(names))
	for _, name := range names {
		tag, err := a.Backend.GetOrCreateTag(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("resolving tag %q: %w", name, err)
		}
		ids = append(ids, tag.ID)
	}
	return ids, nil
}
