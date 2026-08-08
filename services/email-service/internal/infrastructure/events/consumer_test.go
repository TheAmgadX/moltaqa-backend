package events

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestDecodeNotification(t *testing.T) {
	payload, err := structpb.NewStruct(map[string]any{
		"email":  "person@example.com",
		"phone":  "+201234567890",
		"otp":    "123456",
		"action": "Login",
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
	encoded, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	notification, err := decodeNotification(encoded)
	if err != nil {
		t.Fatalf("decodeNotification() error = %v", err)
	}
	if notification.Email != "person@example.com" || notification.Phone != "+201234567890" || notification.OTP != "123456" || notification.Action != "Login" {
		t.Fatalf("unexpected notification: %+v", notification)
	}
}

func TestDecodeNotificationRejectsInvalidPayload(t *testing.T) {
	if _, err := decodeNotification([]byte("not protobuf")); err == nil {
		t.Fatal("decodeNotification() error = nil, want error")
	}
}
