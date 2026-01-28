package auth

import (
	"time"
	"videocall/internal/infrastructure/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Secret []byte
	Ttl    time.Duration
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	RoomID   string `json:"room"`
	jwt.RegisteredClaims
}

func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		Secret: []byte(cfg.JWT.Secret),
		Ttl:    cfg.JWT.TTL,
	}
}

func (j *JWT) Issue(userID, username, roomID string) (string, *jwt.Token, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		userID,
		username,
		roomID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Ttl * time.Second)),
			Issuer:    "test",
		},
	})

	tokenString, err := token.SignedString(j.Secret)

	return tokenString, token, err
}

func (j *JWT) Validate(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return j.Secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
}

func (j *JWT) GetToken(tokenStr string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return j.Secret, nil
	})

	return token, claims, err
}
