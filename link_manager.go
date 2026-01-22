package cascade

import (
	"fmt"

	"gorm.io/gorm"
)

type LinkManager struct {
	db *gorm.DB
}

var ErrShortUrlExists = fmt.Errorf("shortUrl already exists")

func (lm *LinkManager) CreateLink(url string, shortUrl string, accountID uint) error {
	link := &Link{}
	result := lm.db.Where("short_url = ?", shortUrl).First(link)
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
	result = lm.db.Create(link)
	if result.Error != nil {
		return fmt.Errorf("db create failed: %w", result.Error)
	}
	return nil
}

func (lm *LinkManager) GetLink(shortUrl string) (string, error) {
	link := &Link{}
	result := lm.db.Where("short_url = ?", shortUrl).First(link)
	if result.Error != nil {
		return "", fmt.Errorf("db lookup failed: %w", result.Error)
	}
	return link.Url, nil
}
