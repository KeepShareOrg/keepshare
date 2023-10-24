// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Token struct {
	jwt.RegisteredClaims
	UserId    string `json:"userId"`
	ChannelId string `json:"channelId"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

type TokenManager struct {
	TokenSecretKey         string        `mapstructure:"token_secret_key"`
	AccessTokenExpiration  time.Duration `mapstructure:"access_token_expiration"`
	RefreshTokenExpiration time.Duration `mapstructure:"refresh_token_expiration"`
}

func NewTokenManager() (*TokenManager, error) {
	tm := &TokenManager{}
	if err := viper.Unmarshal(tm); err != nil {
		return nil, err
	}
	if tm.TokenSecretKey == "" {
		tm.TokenSecretKey = "000000"
	}
	if tm.AccessTokenExpiration <= 0 {
		tm.AccessTokenExpiration = 2 * time.Hour
	}
	if tm.RefreshTokenExpiration <= 0 {
		tm.RefreshTokenExpiration = 168 * time.Hour
	}
	return tm, nil
}

func (t *TokenManager) GenerateAccessToken(token *Token) (string, error) {
	token.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(t.AccessTokenExpiration))
	ak := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	return ak.SignedString([]byte(t.TokenSecretKey))
}

func (t *TokenManager) GenerateRefreshToken(token *Token) (string, error) {
	token.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(t.RefreshTokenExpiration))
	rk := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	return rk.SignedString([]byte(t.TokenSecretKey))
}

func (t *TokenManager) ValidateToken(tokenString string) (*Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.TokenSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &Token{
			UserId:    claims["userId"].(string),
			ChannelId: claims["channelId"].(string),
			Email:     claims["email"].(string),
			Username:  claims["username"].(string),
		}, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokens generate access token and refresh token
func GenerateTokens(token *Token) (*Tokens, error) {
	tm, err := NewTokenManager()
	if err != nil {
		return nil, err
	}

	accessToken, err := tm.GenerateAccessToken(token)
	if err != nil {
		return nil, err
	}
	refreshToken, err := tm.GenerateRefreshToken(token)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
