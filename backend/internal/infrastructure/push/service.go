package push

import (
	"encoding/json"
	"fmt"
	"log"
	"videocall/internal/domain/repositories"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type Service struct {
	vapidPublicKey  string
	vapidPrivateKey string
	userRepo        repositories.UserRepositoryInterface
}

func NewService(publicKey, privateKey string, userRepo repositories.UserRepositoryInterface) *Service {
	return &Service{
		vapidPublicKey:  publicKey,
		vapidPrivateKey: privateKey,
		userRepo:        userRepo,
	}
}

type NotificationPayload struct {
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Icon  string                 `json:"icon,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func (s *Service) SendNotification(userID string, payload NotificationPayload) error {
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.PushSubscription == nil {
		return fmt.Errorf("user has no push subscription")
	}

	// Convert to webpush subscription format
	sub := &webpush.Subscription{
		Endpoint: user.PushSubscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: user.PushSubscription.Keys.P256dh,
			Auth:   user.PushSubscription.Keys.Auth,
		},
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send notification
	resp, err := webpush.SendNotification(payloadBytes, sub, &webpush.Options{
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             30,
	})

	if err != nil {
		log.Printf("Failed to send push notification to %s (%s): %v", user.Username, userID, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Printf("Push notification failed with status %d for user %s (%s)", resp.StatusCode, user.Username, userID)

		// If subscription is invalid, remove it
		if resp.StatusCode == 404 || resp.StatusCode == 410 {
			_ = s.userRepo.RemovePushSubscription(userID)
		}

		return fmt.Errorf("push failed with status %d", resp.StatusCode)
	}

	log.Printf("✅ Push notification sent to %s (%s)", user.Username, userID)
	return nil
}

func (s *Service) NotifyRoomInvite(inviterUserID, inviterUsername, invitedUserID, roomID string) error {
	return s.SendNotification(invitedUserID, NotificationPayload{
		Title: "Video Call Invitation",
		Body:  fmt.Sprintf("%s invites you to a video call", inviterUsername),
		Icon:  "/logo192.png",
		Data: map[string]interface{}{
			"type":          "room_invite",
			"roomId":        roomID,
			"inviterUserId": inviterUserID,
			"inviterName":   inviterUsername,
		},
	})
}

func (s *Service) NotifyUserJoined(creatorUserID, joinerUsername, roomID string) error {
	return s.SendNotification(creatorUserID, NotificationPayload{
		Title: "Someone joined your room",
		Body:  fmt.Sprintf("%s has joined the video call", joinerUsername),
		Icon:  "/logo192.png",
		Data: map[string]interface{}{
			"type":       "user_joined",
			"roomId":     roomID,
			"joinerName": joinerUsername,
		},
	})
}

func (s *Service) GetPublicKey() string {
	return s.vapidPublicKey
}
