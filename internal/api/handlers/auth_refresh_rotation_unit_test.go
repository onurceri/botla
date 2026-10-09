package handlers

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeResult struct {
	rows    int64
	rowsErr error
}

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rows, r.rowsErr }

type fakeExecer struct {
	result   sql.Result
	err      error
	gotQuery string
	gotArgs  []any
}

func (f *fakeExecer) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.gotQuery = query
	f.gotArgs = args
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestRevokeActiveRefreshToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		execer  *fakeExecer
		wantOK  bool
		wantErr bool
	}{
		{name: "active token is revoked once", execer: &fakeExecer{result: fakeResult{rows: 1}}, wantOK: true},
		{name: "unknown or already revoked token", execer: &fakeExecer{result: fakeResult{rows: 0}}, wantOK: false},
		{name: "unexpected duplicate rows fail closed", execer: &fakeExecer{result: fakeResult{rows: 2}}, wantOK: false},
		{name: "exec error", execer: &fakeExecer{err: errors.New("boom")}, wantErr: true},
		{name: "rows affected error", execer: &fakeExecer{result: fakeResult{rowsErr: errors.New("boom")}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ok, err := revokeActiveRefreshToken(context.Background(), tt.execer, "hash-1")
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, ok)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok)

			// The revocation must be conditional on the token still being active,
			// otherwise concurrent rotations could each succeed.
			q := strings.ToLower(tt.execer.gotQuery)
			assert.Contains(t, q, "update refresh_tokens")
			assert.Contains(t, q, "revoked is not true")
			assert.Equal(t, []any{"hash-1"}, tt.execer.gotArgs)
		})
	}
}
