package host

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var defaultDownloadCache = gocache.New(15*time.Second, 20*time.Second)

// RepositoryProxy caches return values of a Repository.
type RepositoryProxy struct {
	Repository

	downloadCache *gocache.Cache
}

// NewRepositoryProxy returns a new RepositoryProxy.
// If cache is nil, then it uses a default cache that keeps entries for 15 seconds
// and cleans up expired entries every 20 seconds.
func NewRepositoryProxy(wrapped Repository, cache *gocache.Cache) Repository {
	if cache == nil {
		cache = defaultDownloadCache
	}

	return &RepositoryProxy{Repository: wrapped, downloadCache: cache}
}

// GetFile caches the return value of Repository.GetFile.
func (rf *RepositoryProxy) GetFile(fileName string) (string, error) {
	key := rf.Repository.FullName() + fileName
	data, found := rf.downloadCache.Get(key)
	if found {
		return data.(string), nil
	}

	content, err := rf.Repository.GetFile(fileName)
	if err != nil {
		return content, err
	}

	rf.downloadCache.Set(key, content, gocache.DefaultExpiration)
	return content, nil
}
