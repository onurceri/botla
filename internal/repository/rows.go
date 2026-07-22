package repository

import (
	"database/sql"

	"github.com/onurceri/botla-app/pkg/logger"
)

// rowsLogger logs non-fatal rows.Close errors at package level. Most repositories
// don't hold their own logger, so a shared one keeps the close helper simple.
var rowsLogger = logger.New("warn")

// closeRows closes rows and logs any error instead of discarding it. rows.Close
// rarely fails, but silently ignoring it (the previous `_ = rows.Close()`) trips
// errcheck and can hide a connection or transaction-state problem worth surfacing.
func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		rowsLogger.Warn("rows_close_failed", map[string]any{"error": err.Error()})
	}
}
