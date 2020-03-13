package cache

import (
	"encoding/csv"
	"github.com/ksanta/wordofthedaygame/model"
	"io"
	"log"
	"os"
)

type FileCache struct {
	cacheFile string
}

// NewFileCache is a factory method that creates a new FileCache object
func NewFileCache(cacheFile string) Cache {
	return &FileCache{cacheFile}
}

func (cache *FileCache) SetupRequired() bool {
	_, err := os.Stat(cache.cacheFile)
	return os.IsNotExist(err)
}

func (cache *FileCache) CreateCacheWriter() chan model.Word {
	wordChannel := make(chan model.Word)

	go func() {
		file, err := os.Create(cache.cacheFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		csvWriter := csv.NewWriter(file)
		for word := range wordChannel {
			err := csvWriter.Write(word.ToStringSlice())
			if err != nil {
				log.Fatal(err)
			}
		}
		csvWriter.Flush()
	}()

	return wordChannel
}

func (cache *FileCache) LoadWordsFromCache() model.Words {
	var words model.Words
	file, err := os.Open(cache.cacheFile)
	if err != nil {
		log.Fatal(err)
	}
	wordReader := csv.NewReader(file)
	for {
		record, err := wordReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		word := model.NewFromStringSlice(record)
		words = append(words, word)
	}
	return words
}
