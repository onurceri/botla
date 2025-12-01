package db

import (
    "context"
    "database/sql"
)

type User struct {
    ID string
    Email string
    FullName sql.NullString
    AvatarURL sql.NullString
    SubscriptionPlan sql.NullString
}

func GetUserByID(ctx context.Context, pool *sql.DB, id string) (*User, error) {
    var u User
    err := pool.QueryRowContext(ctx, `
        SELECT id, email, full_name, avatar_url, subscription_plan
        FROM users WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
        &u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.SubscriptionPlan,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}
