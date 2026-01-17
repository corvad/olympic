package olympic

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AccountManager struct {
	db               *gorm.DB
	jwtSigningSecret string
}

func InitAccountManager(db *gorm.DB, jwtSigningSecret string) *AccountManager {
	return &AccountManager{db: db, jwtSigningSecret: jwtSigningSecret}
}

func (am *AccountManager) CreateAccount(email string, password string) (*Account, error) {
	existingAccount, err := am.GetAccountByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("db lookup failed: %w", err)
	}

	if existingAccount != nil {
		return nil, fmt.Errorf("Account already exists")
	}

	hashedPassword, err := hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	account := &Account{
		Email:    email,
		Password: hashedPassword,
		Verified: false,
	}

	result := am.db.Create(account)
	if result.Error != nil {
		return nil, result.Error
	}

	return account, nil
}

func (am *AccountManager) GetAccountByEmail(email string) (*Account, error) {
	var account = Account{Email: email}
	result := am.db.First(&account)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &account, nil
}

func (am *AccountManager) Login(email string, password string) (string, string, error) {
	account, err := am.GetAccountByEmail(email)
	if err != nil {
		return "", "", fmt.Errorf("no account found: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return "", "", fmt.Errorf("incorrect password: %w", err)
	}
	/*
	   jwtToken, err := generateJWT(account.ID)

	   	if err != nil {
	   		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	   	}
	*/
	jwtToken, err := am.GenerateJWT(account.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}
	refreshToken := rand.Text()
	hashedRefreshToken, err := hash(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	am.RecordLogin(Login{AccountID: account.ID, IpAddress: "", UserAgent: "", RefreshToken: hashedRefreshToken, ExpiresAt: time.Now().Add(time.Hour * 24 * 7)})

	return refreshToken, jwtToken, nil
}

func (am *AccountManager) RecordLogin(login Login) error {
	result := am.db.Create(&login)
	return result.Error
}

func (am *AccountManager) GenerateJWT(accountID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"account_id": accountID,
		"exp":        time.Now().Add(time.Hour * 1).Unix(),
	})
	jwtToken, err := token.SignedString([]byte(am.jwtSigningSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	return jwtToken, nil
}

func (am *AccountManager) ChangePassword(accountID uint, newPassword string) error {
	bcryptedPassword, err := hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newPassword = bcryptedPassword

	result := am.db.Model(&Account{}).Where("id = ?", accountID).Update("password", newPassword)
	return result.Error
}

func hash(s string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
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
