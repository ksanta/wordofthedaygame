package cache

import "github.com/ksanta/wordofthedaygame/model"

type Cache interface {
	// DoesNotExist returns true if the cache does not exist
	DoesNotExists() bool

	// CreateCacheWriter creates a consumer that listens on the returned channel and writes all WordDetail
	// objects to the cache
	CreateCacheWriter() chan model.WordDetail

	// LoadWordsFromCache loads all the words from the cache
	LoadWordsFromCache() []model.WordDetail
}
