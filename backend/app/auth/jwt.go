package auth

import (
	"time"
	"videocall/app/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Secret []byte
	Ttl    time.Duration
}

type Claims struct {
	Username string `json:"username"`
	Room     string `json:"room"`
	jwt.RegisteredClaims
}

func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		Secret: []byte(cfg.JWT.Secret),
		Ttl:    time.Duration(cfg.JWT.TTL),
	}
}
func (j *JWT) Issue(username, room string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"room":     room,
		"exp":      time.Now().Add(j.Ttl * time.Second).Unix(),
	})

	return token.SignedString(j.Secret)
}

func (j *JWT) Validate(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
}

func (j *JWT) GetToken(tokenStr string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})

	return token, claims, err
}
