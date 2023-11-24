package auth_test

import (
	"sybline/pkg/auth"
	"testing"
	"time"
)

func TestGetConsumerID(t *testing.T) {
	handler := auth.NewSessionHandler()

	// Test scenario where user session doesn't exist
	_, err := handler.GetConsumerID("nonexistent_token", "nonexistent_user")
	if err != auth.ErrNoUserSessionExists {
		t.Errorf("Expected ErrNoUserSessionExists, got %v", err)
	}

	// Test scenario where token has expired
	handler.AddToken("expired_token", "user_with_expired_token", []byte("id"), time.Now().Add(-time.Hour))
	_, err = handler.GetConsumerID("expired_token", "user_with_expired_token")
	if err != auth.ErrNoUserSessionExists {
		t.Errorf("Expected ErrNoUserSessionExists for expired token, got %v", err)
	}

	// Test scenario where valid token exists
	expectedID := []byte("valid_id")
	expiryTime := time.Now().Add(time.Hour)
	handler.AddToken("valid_token", "user_with_valid_token", expectedID, expiryTime)
	id, err := handler.GetConsumerID("valid_token", "user_with_valid_token")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(id) != string(expectedID) {
		t.Errorf("Expected ID %v, got %v", expectedID, id)
	}
}

func TestRemoveToken(t *testing.T) {
	handler := auth.NewSessionHandler()

	// Test scenario where user doesn't exist
	err := handler.RemoveToken("nonexistent_token", "nonexistent_user")
	if err != auth.ErrUserDoesNotExist {
		t.Errorf("Expected ErrNoUserSessionExists, got %v", err)
	}

	// Test scenario where token doesn't exist for user
	handler.AddToken("existing_token", "user_with_existing_token", []byte("id"), time.Now().Add(time.Hour))
	err = handler.RemoveToken("nonexistent_token", "user_with_existing_token")
	if err != auth.ErrTokenDoesNotExist {
		t.Errorf("Expected ErrTokenDoesNotExist, got %v", err)
	}

	// Test scenario where token is successfully removed
	err = handler.RemoveToken("existing_token", "user_with_existing_token")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestAddToken(t *testing.T) {
	handler := auth.NewSessionHandler()

	// Test scenario where token is successfully added
	expectedID := []byte("added_id")
	expiryTime := time.Now().Add(time.Hour)
	err := handler.AddToken("added_token", "user_with_added_token", expectedID, expiryTime)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test scenario where token is added for an existing user
	err = handler.AddToken("added_token_2", "user_with_added_token", []byte("another_id"), expiryTime)
	if err != nil {
		t.Errorf("Expected no error for adding another token to the same user, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	handler := auth.NewSessionHandler()

	// Test scenario where user is successfully deleted
	handler.AddToken("token_to_delete", "user_to_delete", []byte("id"), time.Now().Add(time.Hour))
	err := handler.DeleteUser("user_to_delete")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test scenario where user doesn't exist for deletion
	err = handler.DeleteUser("nonexistent_user")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
