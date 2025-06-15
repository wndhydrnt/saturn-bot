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

func TestCache_DeleteAllByTag(t *testing.T) {
	underTest := setupCache(t)
	err := underTest.SetWithTags("first", []byte("value"), "test")
	require.NoError(t, err, "stores the first item")
	err = underTest.SetWithTags("second", []byte("value"), "test")
	require.NoError(t, err, "stores the second item")
	values, err := underTest.GetAllByTag("test")
	require.NoError(t, err, "reads all the items")
	require.Equal(t, [][]byte{[]byte("value"), []byte("value")}, values)

	err = underTest.DeleteAllByTag("test")
	require.NoError(t, err, "succeeds at deleting all items by tag")
	values, err = underTest.GetAllByTag("test")
	require.NoError(t, err, "reads all the items")
	require.Len(t, values, 0, "no items found")
}

func TestCache_DeleteAllByTag_NoEntries(t *testing.T) {
	underTest := setupCache(t)

	err := underTest.DeleteAllByTag("test")
	require.NoError(t, err, "succeeds at deleting all items by tag")
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

func TestCache_Set_Update(t *testing.T) {
	underTest := setupCache(t)
	err := underTest.Set("unittest", []byte("first"))
	require.NoError(t, err, "stores the first write successfully")
	err = underTest.Set("unittest", []byte("second"))
	require.NoError(t, err, "stores the second write successfully")

	value, err := underTest.Get("unittest")
	require.NoError(t, err, "gets the item")
	require.Equal(t, []byte("second"), value)
}

func TestCache_GetAllByTag(t *testing.T) {
	underTest := setupCache(t)
	values, err := underTest.GetAllByTag("unknown")
	require.NoError(t, err, "succeeds")
	require.Equal(t, 0, len(values))
}

func TestCache_SetWithTag(t *testing.T) {
	underTest := setupCache(t)

	err := underTest.SetWithTags("unittest", []byte("value"), "one", "two")
	require.NoError(t, err, "first set succeeds")

	valuesOneBefore, err := underTest.GetAllByTag("one")
	require.NoError(t, err, "first get by tag 'one' succeeds")
	require.Equal(t, 1, len(valuesOneBefore), "first get by tag 'one' returns the expected number of items")
	require.Equal(t, []byte("value"), valuesOneBefore[0], "first get by tag 'one' returns the expected value")

	valuesTwoBefore, err := underTest.GetAllByTag("two")
	require.NoError(t, err, "first get by tag 'two' succeeds")
	require.Equal(t, 1, len(valuesTwoBefore), "first get by tag 'two' returns the expected number of items")
	require.Equal(t, []byte("value"), valuesTwoBefore[0], "first get by tag 'two' returns the expected value")

	err = underTest.SetWithTags("unittest", []byte("other"), "two")
	require.NoError(t, err, "second set succeeds")

	valuesOneAfter, err := underTest.GetAllByTag("one")
	require.NoError(t, err, "second get by tag 'one' succeeds")
	require.Len(t, valuesOneAfter, 0, "first get by tag 'one' returns no items")

	valuesTwoAfter, err := underTest.GetAllByTag("two")
	require.NoError(t, err, "second get by tag 'two' succeeds")
	require.Equal(t, 1, len(valuesTwoAfter), "second get by tag 'two' returns the expected number of items")
	require.Equal(t, []byte("other"), valuesTwoAfter[0], "second get by tag 'two' returns the expected value")
}
