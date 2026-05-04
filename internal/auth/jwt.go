package auth

import (
	"errors"
	"log"
	"os"
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

// JWT载荷
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// 生成JWT令牌
func GenerateToken(userID, role string) (string, error) {
	exp := time.Now().Add(24 * time.Hour) // 24小时有效期
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// 验证JWT令牌
func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	// 解析 JWT
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
