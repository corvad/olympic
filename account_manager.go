package olympic

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AccountManager struct {
	db *gorm.DB
}

func InitAccountManager(db *gorm.DB) *AccountManager {
	return &AccountManager{db: db}
}

func (am *AccountManager) CreateAccount(email string, password string) (*Account, error) {
	existingAccount, err := am.GetAccountByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("db lookup failed: %w", err)
	}

	if existingAccount != nil {
		return nil, fmt.Errorf("Account already exists")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	account := &Account{
		email:    email,
		password: hashedPassword,
		verified: false,
	}

	result := am.db.Create(account)
	if result.Error != nil {
		return nil, result.Error
	}

	return account, nil
}

func (am *AccountManager) GetAccountByEmail(email string) (*Account, error) {
	var account Account
	result := am.db.Where("email = ?", email).First(&account)
	if result.Error != nil {
		return nil, result.Error
	}
	return &account, nil
}

func (am *AccountManager) VerifyAccount(accountID uint) error {
	result := am.db.Model(&Account{}).Where("id = ?", accountID).Update("verified", true)
	return result.Error
}

func (am *AccountManager) IsAccountVerified(accountID uint) (bool, error) {
	var account Account
	result := am.db.Select("verified").Where("id = ?", accountID).First(&account)
	if result.Error != nil {
		return false, result.Error
	}
	return account.verified, nil
}

func (am *AccountManager) Authenticate(email string, password string) (*Account, error) {
	account, err := am.GetAccountByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	if account.password != password {
		return nil, fmt.Errorf("authentication failed")
	}

	return account, nil
}

func (am *AccountManager) RecordLogin(accountID uint, ipAddress string, userAgent string) error {
	login := &Login{
		accountID: accountID,
		timestamp: gorm.NowFunc(),
		ipAddress: ipAddress,
		userAgent: userAgent,
	}

	result := am.db.Create(login)
	return result.Error
}

func (am *AccountManager) ChangePassword(accountID uint, newPassword string) error {
	bcryptedPassword := bcryptnewPassword

	result := am.db.Model(&Account{}).Where("id = ?", accountID).Update("password", newPassword)
	return result.Error
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password))
	return string(bytes), err
}

func (am *AccountManager) Close() error {
	sqlDB, err := am.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
