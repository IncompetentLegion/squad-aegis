package core

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/guregu/null/v5"

	"github.com/google/uuid"
	"go.codycody31.dev/squad-aegis/internal/db"
	"go.codycody31.dev/squad-aegis/internal/models"
)

func NewSessionToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func CreateSession(ctx context.Context, database db.Executor, userId uuid.UUID, userIp string, expiresIn time.Duration) (*models.Session, error) {
	token, err := NewSessionToken()
	if err != nil {
		return nil, err
	}

	tokenHash := HashSessionToken(token)
	session := &models.Session{
		Id:         uuid.New(),
		UserId:     userId,
		Token:      token,
		LastSeen:   time.Now(),
		LastSeenIp: userIp,
	}
	_, err = database.ExecContext(ctx, "INSERT INTO sessions (id, user_id, token, last_seen, last_seen_ip, created_at) VALUES ($1, $2, $3, $4, $5, $6)", session.Id, session.UserId, tokenHash, session.LastSeen, session.LastSeenIp, session.LastSeen)
	if err != nil {
		return nil, err
	}

	// If expiresAt is not duration of 0, set the expiration time
	if expiresIn != 0 {
		_, err = database.ExecContext(ctx, "UPDATE sessions SET expires_at = $1 WHERE token = $2", time.Now().Add(expiresIn), tokenHash)
		if err != nil {
			return nil, err
		}
		session.ExpiresAt = null.TimeFrom(time.Now().Add(expiresIn))
	}

	return session, nil
}

func GetSessionsByUserId(ctx context.Context, database db.Executor, userId uuid.UUID) ([]models.Session, error) {
	rows, err := database.QueryContext(ctx, "SELECT id, user_id, token, expires_at, last_seen, last_seen_ip FROM sessions WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		if err := rows.Scan(&session.Id, &session.UserId, &session.Token, &session.ExpiresAt, &session.LastSeen, &session.LastSeenIp); err != nil {
			fmt.Println(err)
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func GetSessionById(ctx context.Context, database db.Executor, sessionId uuid.UUID) (*models.Session, error) {
	row := database.QueryRowContext(ctx, "SELECT id, user_id, token, expires_at, last_seen, last_seen_ip FROM sessions WHERE id = $1", sessionId)
	var session models.Session
	if err := row.Scan(&session.Id, &session.UserId, &session.Token, &session.ExpiresAt, &session.LastSeen, &session.LastSeenIp); err != nil {
		return nil, err
	}

	return &session, nil
}

func DeleteSessionById(ctx context.Context, database db.Executor, sessionId uuid.UUID) error {
	_, err := database.ExecContext(ctx, "DELETE FROM sessions WHERE id = $1", sessionId)
	return err
}

func DeleteSessionsByUserIdExcept(ctx context.Context, database db.Executor, userId uuid.UUID, keepSessionId uuid.UUID) error {
	_, err := database.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1 AND id != $2", userId, keepSessionId)
	return err
}
