package usecase

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *ApiUseCases) HandleTurn(w http.ResponseWriter, r *http.Request) {
	// Проверка JWT
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	tok, err := s.jwt.Validate(tokenStr)
	if err != nil || !tok.Valid {
		log.Printf("Invalid JWT token: %v", err)
		http.Error(w, "invalid jwt", http.StatusUnauthorized)
		return
	}

	// Генерация TURN credentials для use-auth-secret
	// TTL 24 часа для стабильности
	ttl := time.Duration(s.cfg.Turn.TTL) * time.Second
	exp := time.Now().Add(ttl)
	timestamp := exp.Unix()

	// КРИТИЧЕСКИ ВАЖНО: username должен быть в формате "timestamp:username"
	// где timestamp - это Unix время истечения срока действия
	username := fmt.Sprintf("%d:webrtc-user", timestamp)

	// Генерация пароля: base64(hmac-sha1(secret, username))
	h := hmac.New(sha1.New, []byte(s.cfg.Turn.Secret))
	h.Write([]byte(username))
	password := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Формирование полных TURN URIs
	turnHost := s.cfg.Turn.Host

	// Убедимся, что используется правильный порт
	// Для TURN без TLS используем порт 3478, для TURNS (TLS) - 5349
	if !strings.Contains(turnHost, ":") {
		turnHost = turnHost + ":3478"
	}

	// Создаем версию хоста для STUN (обычно тот же адрес)
	stunHost := turnHost

	// Список ICE серверов
	uris := []string{
		// TURN UDP (приоритетный для WebRTC)
		fmt.Sprintf("turn:%s?transport=udp", turnHost),
		// TURN TCP (fallback для сетей с блокировкой UDP)
		fmt.Sprintf("turn:%s?transport=tcp", turnHost),
		// STUN сервер (для определения публичного IP)
		fmt.Sprintf("stun:%s", stunHost),
	}

	// Добавляем публичные STUN серверы как fallback
	//publicStun := []string{
	//	"stun:stun.l.google.com:19302",
	//	"stun:stun1.l.google.com:19302",
	//}
	//uris = append(uris, publicStun...)

	creds := map[string]interface{}{
		"username": username,
		"password": password,
		"ttl":      int(ttl.Seconds()),
		"uris":     uris,
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, creds)
}
