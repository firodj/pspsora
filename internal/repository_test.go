package internal

import (
	"context"
	"testing"

	"github.com/firodj/pspsora/models"
	"github.com/stretchr/testify/assert"
)

func TestAka(t *testing.T) {
	repo := NewSQLRepository()

	assert.NotNil(t, repo)
	ctx := context.Background()

	_, err := repo.db.ExecContext(ctx, "SELECT 1")
	assert.NoError(t, err)

	_, err = repo.db.NewCreateTable().Model((*models.BasicBlock)(nil)).Exec(ctx)
	assert.NoError(t, err)
}
