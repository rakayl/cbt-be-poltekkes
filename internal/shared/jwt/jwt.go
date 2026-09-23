package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	AdminSecretKey       = []byte("POLTEKKES_CAT_SECRET_KEY_ADMIN_2026_ENTERPRISE")
	ParticipantSecretKey = []byte("POLTEKKES_CAT_SECRET_KEY_PARTICIPANT_2026_ENTERPRISE")
)

type AdminClaims struct {
	UserID   int      `json:"user_id"`
	Username string   `json:"username"`
	FullName string   `json:"full_name"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type ParticipantClaims struct {
	ParticipantCode string `json:"participant_code"`
	Name            string `json:"name"`
	TokenLogin      string `json:"token_login"`
	jwt.RegisteredClaims
}

func GenerateAdminToken(userID int, username, fullName string, roles []string, duration time.Duration) (string, error) {
	claims := AdminClaims{
		UserID:   userID,
		Username: username,
		FullName: fullName,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "poltekkes-cat-auth",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AdminSecretKey)
}

func ValidateAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return AdminSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid admin access token")
}

func GenerateParticipantToken(participantCode, name, tokenLogin string, duration time.Duration) (string, error) {
	claims := ParticipantClaims{
		ParticipantCode: participantCode,
		Name:            name,
		TokenLogin:      tokenLogin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "poltekkes-cat-front",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ParticipantSecretKey)
}

func ValidateParticipantToken(tokenString string) (*ParticipantClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ParticipantClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ParticipantSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*ParticipantClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid participant session token")
}
