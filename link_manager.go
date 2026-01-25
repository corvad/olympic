package cascade

import (
	"fmt"

	"gorm.io/gorm"
)

type LinkManager struct {
	db *DB
	kv *KVStore
}

var ErrShortUrlExists = fmt.Errorf("shortUrl already exists")

func NewLinkManager(db *DB, kv *KVStore) *LinkManager {
	return &LinkManager{db: db, kv: kv}
}

func (lm *LinkManager) CreateLink(url string, shortUrl string, accountID uint) error {
	link := &Link{}
	result := lm.db.DB.Where("short_url = ?", shortUrl).First(link)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("db lookup failed: %w", result.Error)
	}
	if result.Error == nil {
		return ErrShortUrlExists
	}
	link.Url = url
	link.Owner = accountID
	link.Count = 0
	link.ShortUrl = shortUrl
	result = lm.db.DB.Create(link)
	if result.Error != nil {
		return fmt.Errorf("db create failed: %w", result.Error)
	}
	return nil
}

func (lm *LinkManager) GetLink(shortUrl string) (string, error) {
	link := &Link{}
	result := lm.db.DB.Where("short_url = ?", shortUrl).First(link)
	if result.Error != nil {
		return "", fmt.Errorf("db lookup failed: %w", result.Error)
	}
	return link.Url, nil
}
