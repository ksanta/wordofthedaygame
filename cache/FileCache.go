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

func (cache *FileCache) DoesNotExists() bool {
	_, err := os.Stat(cache.cacheFile)
	return os.IsNotExist(err)
}

func (cache *FileCache) CreateCacheWriter() chan model.WordDetail {
	wordChannel := make(chan model.WordDetail)

	go func() {
		file, err := os.Create(cache.cacheFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		csvWriter := csv.NewWriter(file)
		for details := range wordChannel {
			err := csvWriter.Write(details.ToStringSlice())
			if err != nil {
				log.Fatal(err)
			}
		}
		csvWriter.Flush()
	}()

	return wordChannel
}

func (cache *FileCache) LoadWordsFromCache() []model.WordDetail {
	var allDetails []model.WordDetail
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
		details := model.NewFromStringSlice(record)
		allDetails = append(allDetails, details)
	}
	return allDetails
}
