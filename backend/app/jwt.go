package app

import (
	"time"
	"videocall/app/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		secret: []byte(cfg.JWT.Secret),
		ttl:    time.Duration(cfg.JWT.TTL),
	}
}
func (j *JWT) Issue(username, room string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"room":     room,
		"exp":      time.Now().Add(j.ttl * time.Second).Unix(),
	})

	return token.SignedString(j.secret)
}

func (j *JWT) Validate(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
}
