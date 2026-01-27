
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

#### 42 CREDENTIALS

1. Generate a new API application on the [42 intranet](https://profile.intra.42.fr/oauth/applications/new)
2. In the field Redirect URI add: http://localhost:8080/oauth/callback
3. From the available scopes, choose "Access the user public data" and then proceed to submit.

- Environment variables:
- `CLIENT_ID`: 42 API client ID
- `SECRET`: 42 API client secret
- `REDIRECT_URI`: http://localhost:8080/oauth/callback
- `OAUTH_AUTH_URI`: https://api.intra.42.fr/oauth/authorize
- `OAUTH_TOKEN_URI`: https://api.intra.42.fr/oauth/token
- `JWT_SECRET`: YOUR JWT_SECRET
- `REDIRECT_TOKEN_URI`: http://localhost:8080/?token=

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

1. Create a new API application on the **42 Intranet**
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

## Database Migrations

Run migrations:

```bash
goose postgres "postgres://{username}:{password}b@localhost:5432/bookme" up
```

Or manually apply SQL files from:

```
sql/schema/
```
---

## Run the Application ▶️

```bash
go run .
```

Or:

```bash
go run cmd/api/main.go
```

The server will start at:

```
http://localhost:8080
```