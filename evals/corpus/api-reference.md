# Halvane REST API (v3)

The base URL for the current API is https://api.halvane.example/v3.

## Authentication

Requests are authenticated with API keys. Keys start with the prefix hvk_ and are created by organization administrators in the Console under Settings > API keys. Send the key as a Bearer token in the Authorization header.

## Rate limits

- Business: 600 requests per minute per organization.
- Enterprise: 3,000 requests per minute per organization.

Requests over the limit receive HTTP 429 with a Retry-After header.

## Main endpoints

- GET /sensors lists the sensors in your organization.
- GET /sensors/{id}/readings returns readings for a time range. A single request can cover at most 31 days.
- POST /webhooks registers a webhook.

## Webhooks

Supported events are reading.threshold_exceeded and device.offline. The device.offline event fires when a sensor has sent no data for 30 minutes. Failed webhook deliveries are retried up to 5 times with exponential backoff. Every delivery is signed with HMAC-SHA256 and the signature is sent in the X-Halvane-Signature header.

## Deprecation

API v2 is deprecated and will be shut down (sunset) on 31 December 2025. Migrate integrations to v3 before that date.
