# KVKK Compliance Guide

This document outlines how Botla complies with the Personal Data Protection Law (KVKK) of Turkey and provides instructions for users and administrators on managing personal data.

## 1. Data Collection & Processing

Botla collects and processes the following categories of personal data:

- **Identity Information**: Name, Surname.
- **Contact Information**: Email address.
- **Transaction Security**: IP address, User Agent, Login logs.
- **Customer Transaction**: Chat logs, usage data, subscription details.
- **Marketing**: Consent records for marketing communications.

## 2. User Rights

Under KVKK Article 11, users have the right to:

1. Learn whether their personal data is processed.
2. Request information if their personal data has been processed.
3. Learn the purpose of processing and whether it is used appropriately.
4. Request correction of incomplete or incorrect data.
5. Request deletion or destruction of personal data (Right to be Forgotten).
6. Request notification of corrections/deletions to third parties.
7. Object to results solely from automated systems.
8. Request compensation for damages due to unlawful processing.

## 3. Managing Privacy Settings (For Users)

Users can manage their privacy settings through the "Privacy Settings" page in the dashboard.

### 3.1 Consent Management
Users can view and update their consents for:
- **Marketing**: Promotional emails and notifications.
- **Analytics**: Usage tracking for product improvement.
- **Personalization**: Tailored content and recommendations.
- **Third Party**: Sharing data with partners (if applicable).

### 3.2 Data Export (Portability)
Users can request a copy of all their personal data stored by Botla.
1. Go to **Settings > Privacy**.
2. Click **Request Data Export**.
3. Once processed (usually within 24 hours), a JSON file will be available for download.
4. The download link is valid for 7 days.

### 3.3 Account Deletion
Users can request the deletion of their account and all associated data.
1. Go to **Settings > Privacy**.
2. Click **Delete Account**.
3. Provide a reason (optional) and confirm.
4. The request will be reviewed by an administrator. Upon approval, data is anonymized or soft-deleted immediately.

## 4. Admin Procedures (For Administrators)

Platform administrators manage privacy requests via the **Admin Dashboard > Privacy Requests** page.

### 4.1 Processing Export Requests
1. Navigate to **Privacy Requests**.
2. Filter by status `Pending` and type `Export`.
3. Review the request details.
4. Click **Approve** to trigger the automated export generation.
5. The system will generate the JSON file and notify the user (via the UI status).

### 4.2 Processing Deletion Requests
1. Navigate to **Privacy Requests**.
2. Filter by status `Pending` and type `Deletion`.
3. Review the request. Ensure there are no outstanding billing issues or legal holds.
4. Click **Approve** to execute the deletion.
   - **Action**: The user record is marked as deleted (`deleted_at` is set).
   - **Effect**: The user can no longer log in. Personal data is removed from active processing.
5. Click **Deny** if the request cannot be fulfilled (e.g., legal obligation), providing a reason.

### 4.3 Generating Exports Manually
Admins can generate a data export for any user manually:
1. Go to **Users** list.
2. Select a user.
3. Click **Actions > Export Data**.

## 5. Data Retention Policy

Botla implements automated data retention policies to ensure data is not kept longer than necessary.

- **Retention Period**: 2 Years (default) for all user data.
- **Automated Job**: A daily background job runs at 03:00 AM system time.
- **Scope**:
  - Chat logs
  - Access logs
  - Inactive user accounts (soft-deleted)
- **Action**: Data older than the retention period is permanently deleted (hard delete) or anonymized.

## 6. Technical Implementation Details

- **Database**: `privacy_requests`, `user_consents`, `data_exports` tables.
- **Service**: `PrivacyService` handles logic.
- **Storage**: Export files are stored in secure object storage (R2/S3) with 7-day expiration.
- **Audit**: All admin actions regarding privacy requests are logged in `audit_logs`.
