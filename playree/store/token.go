package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
)

type TokenStore interface {
	Save(ctx context.Context, userID string, token *oauth2.Token) error
	Get(ctx context.Context, userID string) (*oauth2.Token, error)
	Update(ctx context.Context, userID string, newToken *oauth2.Token) error
	Delete(ctx context.Context, userID string) error
}

type tokenStore struct {
	client *redis.Client
	prefix string
}

func NewTokenStore(client *redis.Client, prefix string) TokenStore {
	return &tokenStore{
		client: client,
		prefix: prefix,
	}
}

func (ts *tokenStore) Save(ctx context.Context, userID string, token *oauth2.Token) error {
	key := fmt.Sprintf("%s:%s", ts.prefix, userID)

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	err = ts.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

func (ts *tokenStore) Get(ctx context.Context, userID string) (*oauth2.Token, error) {
	key := fmt.Sprintf("%s:%s", ts.prefix, userID)
	result, err := ts.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Token not found
	} else if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	var token oauth2.Token
	err = json.Unmarshal([]byte(result), &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

func (ts *tokenStore) Delete(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s:%s", ts.prefix, userID)
	err := ts.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil { // Ignore error if key not found
		return fmt.Errorf("failed to delete token:  %w", err)
	}
	return nil
}

func (ts *tokenStore) Update(ctx context.Context, userID string, newToken *oauth2.Token) error {
	existingToken, err := ts.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get existing token: %w", err)
	}

	// If the existing token is not found, consider it an error (optional logic)
	if existingToken == nil {
		return fmt.Errorf("token to update not found")
	}

	return ts.Save(ctx, userID, newToken)
}
