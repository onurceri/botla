package repository

import "encoding/json"

// jsonbString marshals v to a JSON string for storage in a jsonb column.
//
// A string is returned rather than []byte because pgx's simple query protocol
// (default_query_exec_mode=simple_protocol, used to avoid prepared-statement name
// collisions with connection poolers) encodes []byte as PostgreSQL bytea, which is
// rejected for json/jsonb columns with "invalid input syntax for type json". A
// string is parsed as jsonb under every pgx query exec mode. Callers should pass a
// nil interface{} (SQL NULL) separately when the field is absent, since an empty
// string is not valid JSON.
func jsonbString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
