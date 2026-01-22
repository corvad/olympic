package cascade

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrExpiredJWT = fmt.Errorf("JWT has expired")

type AccountManager struct {
	db               *gorm.DB
	jwtSigningSecret string
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
	jwtToken, err := am.GenerateJWT(account.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}
	refreshToken := fmt.Sprintf("%d:%s", account.ID, rand.Text())
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

func (am *AccountManager) ValidateJWT(jwtToken string) (uint, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (any, error) {
		return []byte(am.jwtSigningSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return 0, fmt.Errorf("failed to parse JWT: %w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		exp, ok := claims["exp"].(float64)
		if !ok {
			return 0, fmt.Errorf("invalid exp in JWT claims")
		}
		if exp <= float64(time.Now().Unix()) {
			return 0, ErrExpiredJWT
		}
		accountIDFloat, ok := claims["account_id"].(float64)
		if !ok {
			return 0, fmt.Errorf("invalid account_id in JWT claims")
		}
		return uint(accountIDFloat), nil
	} else {
		return 0, fmt.Errorf("invalid JWT claims")
	}
}

func (am *AccountManager) ValidateRefreshToken(refreshToken string) (Login, error) {
	var logins []Login
	parts := strings.Split(refreshToken, ":")
	if len(parts) != 2 {
		return Login{}, fmt.Errorf("invalid refresh token format")
	}

	accountID, err := strconv.Atoi(parts[0])
	if err != nil {
		return Login{}, fmt.Errorf("invalid account id in token")
	}

	result := am.db.Where("account_id = ? AND expires_at > ?", accountID, time.Now()).Find(&logins)
	if result.Error != nil {
		return Login{}, fmt.Errorf("db query failed: %w", result.Error)
	}

	for _, login := range logins {
		err = bcrypt.CompareHashAndPassword([]byte(login.RefreshToken), []byte(refreshToken))
		if err == nil {
			return login, nil
		}
	}

	return Login{}, fmt.Errorf("no valid refresh token found")
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

func (am *AccountManager) Logout(login Login) error {
	result := am.db.Delete(&Login{}, login.ID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete login: %w", result.Error)
	}
	return nil
}
