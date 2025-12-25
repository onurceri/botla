# Admin Runbook

This guide provides operational procedures and troubleshooting steps for Botla platform administrators.

## 1. Access & Authentication

### 1.1 Accessing the Admin Dashboard
- **URL**: `https://app.botla.co/admin` (or `http://localhost:5173/admin` in dev)
- **Requirements**: User must have `is_platform_admin = true` in the database.

### 1.2 Managing Admin Access
To grant admin privileges to a user (CLI or SQL):

**SQL:**
```sql
UPDATE users SET is_platform_admin = true WHERE email = 'user@example.com';
```

## 2. Monitoring System Health

### 2.1 Health Status
Check the **System Health** page (`/admin/health`) for real-time status of:
- **Database**: PostgreSQL connection.
- **Redis**: Cache and Queue connectivity.
- **Vector DB**: Qdrant availability.
- **External APIs**: OpenAI, etc.

### 2.2 Critical Alerts
If the system is reporting "Unhealthy":
1. Check logs: `docker logs botla-backend`
2. Verify database connectivity.
3. Check disk space on the server.

## 3. Queue Management

### 3.1 Stuck Jobs
Jobs (scraping, embedding) may get stuck in `processing` state if the worker crashes.
1. Go to **Queues > Stuck Jobs**.
2. Review jobs pending for > 30 minutes.
3. Click **Retry** to reset them to `pending`.
4. If a job fails repeatedly, check the "Error Message" and consider deleting the source if the URL is invalid.

### 3.2 High Latency
If queue depth is high (> 1000):
1. Check if we are hitting rate limits with OpenAI or Qdrant.
2. Consider scaling up worker instances.

## 4. User & Organization Management

### 4.1 Suspending a User
If a user is abusive or violating terms:
1. Go to **Users**.
2. Find the user.
3. Click **Suspend**.
4. This immediately revokes API access and login.

### 4.2 Handling Privacy Requests
See [KVKK Compliance Guide](kvkk_compliance.md) for details on Export and Deletion requests.

## 5. Troubleshooting Common Issues

### 5.1 "Chatbot not replying"
- **Cause**: Empty vector store or LLM API error.
- **Fix**: 
  1. Check Chatbot status in Admin Dashboard.
  2. Verify it has "Active" sources with > 0 chunks.
  3. Check **Errors** log for that chatbot.

### 5.2 "Source failing to scrape"
- **Cause**: Website blocking bot, timeout, or invalid selector.
- **Fix**:
  1. Check the error message in **Data Sources**.
  2. If 403 Forbidden, the site may block scrapers.
  3. Suggest user to upload PDF/Text instead.

## 6. Maintenance

### 6.1 Database Migrations
Migrations are run automatically on deployment. To run manually:
```bash
make migrate-up
```

### 6.2 Backups
- Database is backed up daily to S3.
- To restore: Use the `pg_restore` tool with the latest dump.
