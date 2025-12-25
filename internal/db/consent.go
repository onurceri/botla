package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UserConsent struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	ConsentType string     `json:"consent_type"`
	Granted     bool       `json:"granted"`
	GrantedAt   time.Time  `json:"granted_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

func GetUserConsents(ctx context.Context, pool *sql.DB, userID string) ([]UserConsent, error) {
	rows, err := pool.QueryContext(ctx, `
		SELECT id, user_id, consent_type, granted, granted_at, revoked_at
		FROM user_consents
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user consents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var consents []UserConsent
	for rows.Next() {
		var c UserConsent
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ConsentType, &c.Granted, &c.GrantedAt, &c.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user consent: %w", err)
		}
		consents = append(consents, c)
	}
	return consents, nil
}

func UpsertConsent(ctx context.Context, pool *sql.DB, userID, consentType string, granted bool, ip, userAgent string) error {
	var revokedAt *time.Time
	if !granted {
		now := time.Now()
		revokedAt = &now
	}

	_, err := pool.ExecContext(ctx, `
		INSERT INTO user_consents (user_id, consent_type, granted, ip_address, user_agent, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, consent_type) 
		DO UPDATE SET 
			granted = EXCLUDED.granted,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent,
			revoked_at = EXCLUDED.revoked_at,
			granted_at = CASE WHEN EXCLUDED.granted THEN NOW() ELSE user_consents.granted_at END
	`, userID, consentType, granted, ip, userAgent, revokedAt)

	if err != nil {
		return fmt.Errorf("upsert consent: %w", err)
	}
	return nil
}
