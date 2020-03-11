package cache

import "github.com/ksanta/wordofthedaygame/model"

type Cache interface {
	// DoesNotExist returns true if the cache does not exist
	DoesNotExist() bool

	// CreateCacheWriter creates a consumer that listens on the returned channel and persists all words sent to the
	// channel to the cache
	CreateCacheWriter() chan model.Word

	// LoadWordsFromCache loads all the words from the cache
	LoadWordsFromCache() model.Words
}
