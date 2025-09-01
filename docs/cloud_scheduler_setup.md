# Setting Up Cloud Scheduler for Daily Updates

This guide explains how to set up Google Cloud Scheduler to automatically update your Yoto card daily with a new bird song.

## Prerequisites

1. Your Bird Song Explorer service is deployed to Cloud Run
2. You have obtained and set the Yoto access/refresh tokens
3. You have set the `YOTO_CARD_ID` environment variable with your MYO card ID

## Step 1: Enable Cloud Scheduler API

```bash
gcloud services enable cloudscheduler.googleapis.com
```

## Step 2: Create a Service Account (Optional but Recommended)

Create a dedicated service account for Cloud Scheduler:

```bash
gcloud iam service-accounts create bird-song-scheduler \
    --display-name="Bird Song Scheduler"

# Grant permission to invoke Cloud Run
gcloud run services add-iam-policy-binding yoto-bird-song-explorer \
    --member="serviceAccount:bird-song-scheduler@yoto-bird-song-explorer.iam.gserviceaccount.com" \
    --role="roles/run.invoker" \
    --region=us-central1
```

## Step 3: Create the Scheduled Job

Create a Cloud Scheduler job that runs daily at 6 AM Pacific Time:

```bash
gcloud scheduler jobs create http daily-bird-update \
    --location=us-central1 \
    --schedule="0 6 * * *" \
    --time-zone="America/Los_Angeles" \
    --uri="https://yoto-bird-song-explorer-[YOUR-PROJECT-ID].a.run.app/api/v1/daily-update" \
    --http-method=POST \
    --oidc-service-account-email="bird-song-scheduler@yoto-bird-song-explorer.iam.gserviceaccount.com" \
    --attempt-deadline="10m" \
    --headers="Content-Type=application/json" \
    --message-body="{}"
```

Replace `[YOUR-PROJECT-ID]` with your actual Cloud Run service URL.

### Schedule Format

The schedule uses cron format:
- `"0 6 * * *"` = Every day at 6:00 AM
- `"0 7 * * *"` = Every day at 7:00 AM
- `"30 6 * * *"` = Every day at 6:30 AM

### Time Zones

Common time zones:
- `America/Los_Angeles` - Pacific Time
- `America/Denver` - Mountain Time (Bend, OR)
- `America/Chicago` - Central Time
- `America/New_York` - Eastern Time

## Step 4: Add Security Token (Optional)

For extra security, you can add a secret token that Cloud Scheduler must provide:

1. Generate a random token:
```bash
openssl rand -hex 32
```

2. Set it as an environment variable in Cloud Run:
```bash
gcloud run services update yoto-bird-song-explorer \
    --update-env-vars SCHEDULER_TOKEN=your_generated_token \
    --region=us-central1
```

3. Update the scheduler job to include the token:
```bash
gcloud scheduler jobs update http daily-bird-update \
    --location=us-central1 \
    --headers="Content-Type=application/json,X-Scheduler-Token=your_generated_token"
```

## Step 5: Test the Scheduled Job

You can manually trigger the job to test it:

```bash
gcloud scheduler jobs run daily-bird-update --location=us-central1
```

Check the logs:

```bash
gcloud logging read "resource.type=cloud_scheduler_job AND resource.labels.job_id=daily-bird-update" \
    --limit=10 \
    --format=json
```

## Step 6: Monitor the Job

View job status:

```bash
gcloud scheduler jobs describe daily-bird-update --location=us-central1
```

List all scheduled jobs:

```bash
gcloud scheduler jobs list --location=us-central1
```

## Troubleshooting

### Check Cloud Run Logs

```bash
gcloud run logs read --service=yoto-bird-song-explorer --region=us-central1 --limit=50
```

### Update Job Schedule

```bash
gcloud scheduler jobs update http daily-bird-update \
    --location=us-central1 \
    --schedule="0 7 * * *"  # Change to 7 AM
```

### Pause/Resume Job

```bash
# Pause
gcloud scheduler jobs pause daily-bird-update --location=us-central1

# Resume
gcloud scheduler jobs resume daily-bird-update --location=us-central1
```

### Delete Job

```bash
gcloud scheduler jobs delete daily-bird-update --location=us-central1
```

## How It Works

1. Every day at the scheduled time, Cloud Scheduler sends a POST request to your `/api/v1/daily-update` endpoint
2. The endpoint:
   - Selects a new bird for the day based on your location (Bend, OR)
   - Gets the bird's song from Xeno-canto
   - Selects a random intro audio
   - Uploads both audio files to Yoto
   - Creates a playlist with the intro and bird song
   - Updates your MYO card with the new content
3. Your Yoto player will have fresh content when played that day!

## Cost

Cloud Scheduler offers 3 free jobs per month. After that, it's $0.10 per job per month. Running one daily job will be free!