package main

import (
	"context"
	"fmt"
	"time"

	"github.com/uh-kay/glimpze/store"
)

func (app *application) updateUserLimits(ctx context.Context) error {
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)

	waitDuration := time.Until(nextRun)

	select {
	case <-time.After(waitDuration):
	case <-ctx.Done():
		app.logger.Info("context cancelled during initial wait")
		return ctx.Err()
	}

	if err := app.batchUpdateUserLimit(ctx); err != nil {
		app.logger.Error("first batch update failed", "error", err)
		return err
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := app.batchUpdateUserLimit(ctx); err != nil {
				app.logger.Error("batch update failed", "error", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (app *application) batchUpdateUserLimit(ctx context.Context) error {
	const (
		limit     = 1000
		batchSize = 100
	)

	offset := 0
	for {
		userIDs, err := app.store.Users.GetIDs(ctx, limit, int64(offset))
		if err != nil {
			return fmt.Errorf("failed to get user IDs: %w", err)
		}
		if len(userIDs) == 0 {
			break
		}

		for i := 0; i < len(userIDs); i += batchSize {
			end := min(i+batchSize, len(userIDs))
			batch := userIDs[i:end]

			if err := app.updateBatch(ctx, batch); err != nil {
				app.logger.Error("batch update failed", "error", err, "batch_size", len(batch))
				continue
			}
		}

		offset += limit
	}
	return nil
}

func (app *application) updateBatch(ctx context.Context, userIDs []int) error {
	for _, userID := range userIDs {
		err := app.store.WithTx(ctx, func(s *store.Storage) error {
			if _, err := app.store.UserLimits.Add(ctx, int64(userID)); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
