package cache_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
)

func setupCache(t *testing.T) *cache.Cache {
	c, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	require.NoError(t, err, "instantiates the cache")
	return c
}

func TestCache_Delete(t *testing.T) {
	underTest := setupCache(t)
	err := underTest.Set("unittest", []byte("value"))
	require.NoError(t, err, "stores the item")
	err = underTest.Delete("unittest")

	require.NoError(t, err, "deletes the item")
	value, err := underTest.Get("unittest")
	require.ErrorIs(t, err, cache.ErrNotFound, "indicates that the item does not exist")
	require.Nil(t, value, "does not find the item")
}

func TestCache_SetGet(t *testing.T) {
	underTest := setupCache(t)
	err := underTest.Set("unittest", []byte("value"))
	require.NoError(t, err, "stores the item")

	value, err := underTest.Get("unittest")
	require.NoError(t, err, "gets the item")
	require.Equal(t, []byte("value"), value)
}

func TestCache_Set_NilValue(t *testing.T) {
	underTest := setupCache(t)
	err := underTest.Set("unittest", nil)
	require.NoError(t, err, "no error")

	value, err := underTest.Get("unittest")
	require.ErrorIs(t, err, cache.ErrNotFound, "indicates that the item does not exist")
	require.Nil(t, value, "does not find the item")
}
