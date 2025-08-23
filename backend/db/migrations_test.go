package db

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestApplyMigrationsIsIdempotent(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, ApplyMigrations(gdb))
	require.NoError(t, ApplyMigrations(gdb))

	has, err := columnExists(gdb, "image_loras", "weight")
	require.NoError(t, err)
	require.True(t, has)

	has, err = columnExists(gdb, "images", "model_id")
	require.NoError(t, err)
	require.True(t, has)

	has, err = columnExists(gdb, "images", "model_name")
	require.NoError(t, err)
	require.False(t, has)

	has, err = columnExists(gdb, "images", "model_hash")
	require.NoError(t, err)
	require.False(t, has)

	has, err = columnExists(gdb, "images", "favorite")
	require.NoError(t, err)
	require.True(t, has)
}
