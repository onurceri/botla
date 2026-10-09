# Data Retention and Privacy

## Retention of sensor readings

- Starter: readings are retained for 13 months.
- Business: readings are retained for 36 months.
- Enterprise: retention is configurable up to 10 years.

Older readings are deleted automatically when they pass the retention period.

## Where data is stored

Halvane stores customer data in Amazon Web Services region eu-central-1 (Frankfurt, Germany). Backups are kept in eu-west-1 (Ireland). Data is encrypted at rest with AES-256 and in transit with TLS 1.2 or newer.

## Account closure and export

Before closing an account, administrators can export all readings from the Console as CSV or Parquet files. After an account is closed, all customer data is deleted within 30 days.

## GDPR

Halvane acts as a data processor under the GDPR, and a Data Processing Agreement (DPA) is available to all customers. Halvane does not sell customer data. The current list of sub-processors is published at halvane.example/subprocessors.
