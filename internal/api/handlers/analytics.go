package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/pkg/middleware"
)

type AnalyticsHandlers struct {
	DB *sql.DB
}

type analyticsPoint struct {
	Date          string `json:"date"`
	Messages      int    `json:"messages"`
	Conversations int    `json:"conversations"`
}

func (h *AnalyticsHandlers) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.QueryContext(r.Context(), `
		WITH dates AS (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '6 days',
				CURRENT_DATE,
				'1 day'::interval
			)::date AS date
		),
		user_analytics AS (
			SELECT a.analytics_date, a.total_messages, a.total_conversations
			FROM analytics a
			JOIN chatbots c ON a.chatbot_id = c.id
			WHERE c.user_id = $1
		)
		SELECT 
			to_char(d.date, 'YYYY-MM-DD') as date,
			COALESCE(SUM(ua.total_messages), 0)::INTEGER as messages,
			COALESCE(SUM(ua.total_conversations), 0)::INTEGER as conversations
		FROM dates d
		LEFT JOIN user_analytics ua ON ua.analytics_date = d.date
		GROUP BY d.date
		ORDER BY d.date
	`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	var data []analyticsPoint
	for rows.Next() {
		var p analyticsPoint
		if errScan := rows.Scan(&p.Date, &p.Messages, &p.Conversations); errScan != nil {
			continue
		}
		data = append(data, p)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		data = []analyticsPoint{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
