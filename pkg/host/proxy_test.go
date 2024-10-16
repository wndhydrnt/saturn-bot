package host_test

import (
	"errors"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
)

type repositoryMock struct {
	host.Repository

	getFileCallCount int
	getFileContent   string
	getFileErr       error
}

func (m *repositoryMock) FullName() string {
	return "git.local/unit/test"
}

func (m *repositoryMock) GetFile(fileName string) (string, error) {
	m.getFileCallCount++
	return m.getFileContent, m.getFileErr
}

func TestRepositoryProxy_GetFile_CacheValid(t *testing.T) {
	mock := &repositoryMock{
		getFileContent: "test",
	}
	proxy := host.NewRepositoryProxy(mock, nil)

	resultOne, err := proxy.GetFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, "test", resultOne)

	resultTwo, err := proxy.GetFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, "test", resultTwo)

	require.Equal(t, 1, mock.getFileCallCount)
}

func TestRepositoryProxy_GetFile_CacheExpired(t *testing.T) {
	mock := &repositoryMock{
		getFileContent: "test",
	}
	proxy := host.NewRepositoryProxy(mock, gocache.New(1*time.Nanosecond, 500*time.Nanosecond))

	resultOne, err := proxy.GetFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, "test", resultOne)

	time.Sleep(200 * time.Nanosecond)

	resultTwo, err := proxy.GetFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, "test", resultTwo)

	require.Equal(t, 2, mock.getFileCallCount)
}

func TestRepositoryProxy_GetFile_Error(t *testing.T) {
	mock := &repositoryMock{
		getFileErr: errors.New("get file failed"),
	}
	proxy := host.NewRepositoryProxy(mock, gocache.New(1*time.Nanosecond, 500*time.Nanosecond))

	_, err := proxy.GetFile("test.txt")
	require.EqualError(t, err, "get file failed")
	require.Equal(t, 1, mock.getFileCallCount)
}
