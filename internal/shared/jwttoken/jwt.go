package jwttoken

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/vault/api"
)

func GetPublicKey(client *api.Client, keyName string) (crypto.PublicKey, error) {
	secret, err := client.Logical().Read("transit/keys/" + keyName)
	if err != nil {
		return nil, err
	}

	keysData, ok := secret.Data["keys"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("could not find keys data")
	}

	var rawKeyB64 string
	maxVer := 0
	for verStr, data := range keysData {
		var ver int
		fmt.Sscanf(verStr, "%d", &ver)
		if ver >= maxVer {
			maxVer = ver
			keyInfo := data.(map[string]any)
			rawKeyB64 = keyInfo["public_key"].(string)
		}
	}

	if rawKeyB64 == "" {
		return nil, fmt.Errorf("public_key is empty")
	}

	// 1. Декодируем Base64 напрямую (без PEM)
	pubBytes, err := base64.StdEncoding.DecodeString(rawKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 public key: %w", err)
	}

	// 2. Для Ed25519 публичный ключ должен быть ровно 32 байта
	if len(pubBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d, want %d", len(pubBytes), ed25519.PublicKeySize)
	}

	return ed25519.PublicKey(pubBytes), nil
}

// ParseToken проверяет подпись и возвращает клеймсы
func ParseToken(tokenString string, publicKey crypto.PublicKey) (*jwt.RegisteredClaims, error) {
	// 1. Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Важно: проверяем, что алгоритм именно Ed25519
		// Если злоумышленник подменит alg на "HS256", библиотека может попытаться проверить RSA-ключ как HMAC-секрет
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга: %w", err)
	}

	// 2. Достаем клеймсы и проверяем валидность
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("невалидный токен")
}
