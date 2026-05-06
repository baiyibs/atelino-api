package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

// 初始化JWT
func InitJWT() {
	secret := os.Getenv("JWT_KEY")
	if secret == "" {
		log.Fatal("没有设置JWT_KEY")
	}
	jwtKey = []byte(secret)
}

// Access Token 载荷
type AccessClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// 生成 Access Token (访问令牌)
func GenerateAccessToken(userID uint, role string) (string, error) {
	exp := time.Now().Add(15 * time.Minute) // 24小时有效期
	claims := AccessClaims{
		UserID: strconv.FormatUint(uint64(userID), 10),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// 验证 Access Token
func ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("无效的 Access Token")
	}
	return claims, nil
}

// 生成 Refresh Token (刷新令牌)
func GenerateRefreshToken() (rawToken string, hash string, err error) {
	// 生成随机字节
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(rawToken))
	hash = hex.EncodeToString(sum[:])
	return rawToken, hash, nil
}

// 对于前端给定的 Refresh Token 计算 Hash
func HashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
