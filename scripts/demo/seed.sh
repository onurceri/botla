#!/usr/bin/env bash
# Seed a local Botla stack with demo data for the README screenshots.
#
# Creates demo@example.com, a "Halvane Instruments" organization (the fictional
# company from evals/corpus), a support bot trained on the 13 corpus documents,
# three smaller bots, and a handful of real conversations, feedback and a
# handoff request through the public widget API. Re-running it deletes the demo
# bots and recreates them.
#
# Needs: the dev stack (`make up`), the API on $API_URL with working embedding
# and chat models (see docs/screenshots.md), curl and jq.
set -euo pipefail

API_URL=${API_URL:-http://localhost:8080}
EMAIL=${DEMO_EMAIL:-demo@example.com}
PASSWORD=${DEMO_PASSWORD:-DemoPass123!}
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
CORPUS="$ROOT/evals/corpus"
JAR=$(mktemp)
trap 'rm -f "$JAR"' EXIT

api() { # api METHOD PATH [curl args...]
  local method=$1 path=$2; shift 2
  # --retry waits out the API's rate limits (HTTP 429).
  curl -sS --fail-with-body --retry 15 --retry-delay 6 -b "$JAR" -c "$JAR" -X "$method" "$API_URL$path" \
    -H "X-Organization-ID: ${ORG_ID:-}" -H "X-Workspace-ID: ${WS_ID:-}" "$@"
}
json() { api "$1" "$2" -H 'Content-Type: application/json' -d "$3"; }

echo "==> demo user"
creds=$(jq -nc --arg e "$EMAIL" --arg p "$PASSWORD" '{email:$e,password:$p,full_name:"Demo User"}')
if ! json POST /api/v1/auth/login "$creds" >/dev/null 2>&1; then
  json POST /api/v1/auth/register "$creds" >/dev/null
  # Dev-only seed step: the demo needs Ultra-plan features (several bots, human
  # handoff). Plans are normally changed by billing or an admin; this sets it
  # directly in the local dev database. Plans are cached in Redis for 5 minutes,
  # which is why it runs right after registration.
  docker exec botla-postgres psql -qU botla -d botla_dev -c \
    "UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'ultra') WHERE email = '$EMAIL'" >/dev/null
  json POST /api/v1/auth/login "$creds" >/dev/null
fi
api POST /api/v1/me/onboarding/skip >/dev/null || true

ORG_ID=$(api GET /api/v1/organizations | jq -r '.[0].id')
WS_ID=$(api GET "/api/v1/organizations/$ORG_ID/workspaces" | jq -r '.[0].id')
json PATCH "/api/v1/organizations/$ORG_ID" '{"name":"Halvane Instruments"}' >/dev/null
json PATCH "/api/v1/organizations/$ORG_ID/workspaces/$WS_ID" '{"name":"Customer Support"}' >/dev/null

echo "==> reset demo bots"
for id in $(api GET /api/v1/chatbots | jq -r 'if type == "array" then .[].id else .chatbots[]?.id end'); do
  api DELETE "/api/v1/chatbots/$id" >/dev/null
done

create_bot() { # create_bot NAME DESCRIPTION COLOR WELCOME -> id
  json POST /api/v1/chatbots "$(jq -nc --arg n "$1" --arg d "$2" --arg c "$3" --arg w "$4" \
    '{name:$n,description:$d,language:"en",model:"gpt-6-luna",theme_color:$c,welcome_message:$w}')" | jq -r '.id'
}
add_text() { # add_text BOT FILE
  api POST "/api/v1/chatbots/$1/sources" -F source_type=text -F "text=<$2" >/dev/null
}

echo "==> bots and sources"
SALES=$(create_bot "Beacon Sales Assistant" "Pre-sales questions on pricing and trials." "#7C3AED" "Planning a rollout? Ask me about pricing.")
add_text "$SALES" "$CORPUS/pricing-and-plans.md"
IT=$(create_bot "Installer Helpdesk" "Helps installers mount and connect sensors." "#EA580C" "Need help installing a Beacon?")
add_text "$IT" "$CORPUS/installation-guide.md"
add_text "$IT" "$CORPUS/troubleshooting-led-codes.md"
PRIV=$(create_bot "Privacy Desk" "Data retention and GDPR questions." "#2563EB" "Questions about your data? I can help.")
add_text "$PRIV" "$CORPUS/data-retention-and-privacy.md"

MAIN=$(create_bot "Halvane Support" "Answers customer questions about Beacon sensors, plans and the Console." \
  "#0F766E" "Hi! I'm the Halvane assistant. Ask me about Beacon sensors, plans or the Console.")
json PUT "/api/v1/chatbots/$MAIN/appearance" '{"theme_color":"#0F766E","chat_header_color":"#0F766E","user_message_color":"#0F766E","send_button_color":"#0F766E","chat_header_text_color":"#FFFFFF","chat_background_color":"#FFFFFF","bot_message_color":"#F1F5F9","bot_display_name":"Halvane Assistant","position":"bottom-right","suggested_questions":["What does the Beacon S2 measure?","How much does the Business plan cost?","What does a blinking amber LED mean?"]}' >/dev/null
json PUT "/api/v1/chatbots/$MAIN/handoff" '{"handoff_enabled":true,"handoff_type":"email"}' >/dev/null
for f in "$CORPUS"/*.md; do add_text "$MAIN" "$f"; done

echo -n "==> waiting for ingestion"
for _ in $(seq 1 90); do
  pending=$(api GET "/api/v1/chatbots/$MAIN/sources" | jq '[(if type == "array" then .[] else .sources[] end) | select(.status != "completed" and .status != "failed")] | length')
  [ "$pending" = 0 ] && break
  echo -n "."; sleep 2
done
echo

chat() { # chat SESSION MESSAGE -> message_id
  curl -sS --fail-with-body --retry 15 --retry-delay 6 -X POST "$API_URL/api/v1/public/chatbots/$MAIN/chat" -H 'Content-Type: application/json' \
    -d "$(jq -nc --arg s "$1" --arg m "$2" '{session_id:$s,message:$m}')" | jq -r '.message_id'
}
feedback() { # feedback MESSAGE_ID true|false
  curl -sS --fail-with-body --retry 15 --retry-delay 6 -o /dev/null -X POST "$API_URL/api/v1/public/chatbots/$MAIN/feedback" \
    -H 'Content-Type: application/json' -d "{\"message_id\":\"$1\",\"thumbs_up\":$2}"
}

echo "==> conversations"
m=$(chat demo-1 "What does the Beacon S2 measure, and how long does the battery last?"); feedback "$m" true
chat demo-1 "Does it work on a 5 GHz Wi-Fi network?" >/dev/null
m=$(chat demo-2 "How much is the Business plan, and do schools get a discount?"); feedback "$m" true
m=$(chat demo-3 "My sensor is blinking amber. What should I do?"); feedback "$m" true
chat demo-4 "Do you integrate with Home Assistant?" >/dev/null
m=$(chat demo-5 "Can I buy Halvane sensors at my local electronics store in Brazil?"); feedback "$m" false
chat demo-6 "We have 400 sensors across 12 schools. Can I talk to someone about an Enterprise quote?" >/dev/null

echo "==> handoff"
req=$(json POST "/api/v1/public/chatbots/$MAIN/handoff" '{"session_id":"demo-6","message":"Enterprise quote for 12 schools"}' | jq -r '.request_id // .id')
json POST "/api/v1/public/chatbots/$MAIN/handoff/$req/contact" '{"email":"j.berg@example.com"}' >/dev/null

echo "Done. Log in at http://localhost:5173 as $EMAIL / $PASSWORD (main bot: $MAIN)"
