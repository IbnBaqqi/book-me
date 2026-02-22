
## Prerequisites

- Go **1.22** or higher  
- PostgreSQL **14+**  
- SMTP server access (Gmail, SendGrid, etc.)  
- **42 Intra OAuth** credentials  

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/IbnBaqqi/book-me.git
cd book-me
```

### 2. Install Dependencies

```bash
go mod download
```

---

## PostgreSQL Configuration 🗄️

Create the database:

```bash
createdb bookme
```

Or using `psql`:

```sql
CREATE DATABASE bookme;
```

---

## Environment Configuration

### 1. Create `.env` file

```bash
cp .env.example .env
```

### 2. Configure Environment Variables

#### Server Configuration
```bash
PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=60s
LOG_LEVEL=info
```

#### App Configuration
```bash
ENV=dev
```

#### Database

```bash
DB_URL=postgres://username:password@localhost:5432/bookme?sslmode=disable
```

#### JWT & Session Secrets

Generate secure secrets:

```bash
openssl rand -base64 32
```

Set them in `.env`:

```bash
JWT_SECRET=your-generated-secret-here
SESSION_SECRET=another-generated-secret-here
```

---

## 42 Intra OAuth Configuration

1. Create a new API application on the **42 Intranet**:  
   https://profile.intra.42.fr/oauth/applications/new
2. Set **Redirect URI**:
   ```
   http://localhost:8080/oauth/callback
   ```
3. Select scope:
   - **Access the user public data**
4. Save and copy your credentials

Add to `.env`:

```bash
CLIENT_ID=your-42-client-id
SECRET=your-42-client-secret
REDIRECT_URI=http://localhost:8080/oauth/callback

OAUTH_AUTH_URI=https://api.intra.42.fr/oauth/authorize
OAUTH_TOKEN_URI=https://api.intra.42.fr/oauth/token
REDIRECT_TOKEN_URI=http://localhost:3000/auth/callback
USER_INFO_URL=https://api.intra.42.fr/v2/me
```

---

## Email Configuration (SMTP)

### Gmail Setup

1. Enable **2-Factor Authentication**
2. Generate an **App Password**  
   https://myaccount.google.com/apppasswords  
3. Select **Mail** and your device
4. Copy the generated password

Add to `.env`:

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-gmail-app-password

FROM_EMAIL=noreply@bookme.com
FROM_NAME=BookMe

SMTP_USE_TLS=true
```

---

## Google Calendar Configuration (Optional - Staff Feature)

Google Calendar integration allows staff bookings to automatically sync with Google Calendar.

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Calendar API**

### 2. Create Service Account

1. Navigate to **IAM & Admin** → **Service Accounts**
2. Click **Create Service Account**
3. Give it a name (e.g., "book-me-calendar")
4. Grant it the **Editor** role
5. Click **Done**

### 3. Generate Service Account Key

1. Click on the created service account
2. Go to **Keys** tab
3. Click **Add Key** → **Create new key**
4. Select **JSON** format
5. Download the JSON file
6. Save it as `assets/book-me-service-account.json`

### 4. Share Calendar with Service Account

1. Open Google Calendar
2. Go to calendar settings
3. Under **Share with specific people**, add the service account email
4. Grant **Make changes to events** permission
5. Copy the **Calendar ID** (found in calendar settings)

### 5. Configure Environment Variables

Add to `.env`:

```bash
GOOGLE_CREDENTIALS_FILE=assets/book-me-service-account.json
GOOGLE_CALENDAR_SCOPE=https://www.googleapis.com/auth/calendar
GOOGLE_CALENDAR_ID=your-calendar-id@group.calendar.google.com
```

---

## Database Migrations

Run migrations using goose:

```bash
goose -dir sql/schema postgres "postgres://username:password@localhost:5432/bookme?sslmode=disable" up
```

Or manually apply SQL files in order from:

```
sql/schema/
  001_users.sql
  002_rooms.sql
  003_reservations.sql
  004_populate_rooms.sql
```
---

## Run the Application ▶️

```bash
make run
```

Or:

```bash
go run cmd/server/main.go
```

The server will start at:

```
http://localhost:8080
```

---

## Available Make Commands

```bash
make run            # Run the application
make build          # Build the binary to bin/book-me
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make clean          # Clean build artifacts
make sqlc           # Generate SQLC code from SQL queries
make fmt            # Format code
make deps           # Install and tidy dependencies
```