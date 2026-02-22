## API Overview

Authentication Flow

1. User clicks "Login" → Redirected to /api/oauth/login
2. System redirects to OAuth provider (42 Intra)
3. User authorizes the application
4. OAuth provider redirects to /oauth/callback
5. System exchanges code for access token
6. System fetches user info and creates/updates user in database
7. System generates JWT token
8. JWT token returned to client
9. Client includes JWT in Authorization: Bearer <token> header for protected routes

## API Endpoints
### Authentication

| Method | Endpoint              | Description                 | Auth Required |
|------|-----------------------|-----------------------------|---------------|
| GET  | /oauth/login          | Initiate OAuth login        | No            |
| GET  | /oauth/callback       | OAuth callback handler      | No            |

### Reservations

| Method | Endpoint                         | Description                         | Auth Required |
|------|----------------------------------|-------------------------------------|---------------|
| POST | /api/v1/reservations             | Create a new reservation            | Yes           |
| GET  | /api/v1/reservations             | Get unavailable time slots          | Yes           |
| DELETE | /api/v1/reservations/{id}      | Cancel a reservation                | Yes           |

### Health Check

| Method | Endpoint        | Description             | Auth Required |
|------|-----------------|-------------------------|---------------|
| GET  | /api/v1/health  | Health check endpoint   | No            |

---

## Health Check Response

The health endpoint checks the status of critical services:

```bash
curl http://localhost:8080/api/v1/health
```

**Response (Healthy)**

```json
{
	"status": "healthy",
	"checks": {
        "calendar": "healthy",
        "database": "healthy",
        "email": "healthy"
    }
}
```

**Response (Unhealthy)**

```json
{
    "status": "unhealthy",
    "checks": {
        "calendar": "healthy",
        "database": "unhealthy",
        "email": "healthy"
    }
}
```

---

## API Examples 📝

### Create a Reservation

```bash
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roomId": 1,
    "startTime": "2025-01-28T14:00:00Z",
    "endTime": "2025-01-28T16:00:00Z"
  }'
```

**Response**

```json
{
  "id": 123,
  "roomId": 1,
  "startTime": "2025-01-28T14:00:00Z",
  "endTime": "2025-01-28T16:00:00Z",
  "createdBy": {
    "id": 42,
    "name": "John Doe"
  }
}
```

---

### Get Unavailable Slots

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/v1/reservations?start=2025-01-28&end=2025-02-01"
```

**Response**

```json
[
  {
    "roomId": 1,
    "roomName": "Big Conference Room",
    "slots": [
      {
        "id": 123,
        "startTime": "2025-01-28T14:00:00Z",
        "endTime": "2025-01-28T16:00:00Z",
        "bookedBy": "John Doe"
      },
	  {
        "id": 124,
        "startTime": "2025-01-29T14:00:00Z",
        "endTime": "2025-01-29T16:00:00Z",
        "bookedBy": "null"
      }
    ]
  }
]
```

---

### Cancel a Reservation

```bash
curl -X DELETE http://localhost:8080/api/v1/reservations/123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response**

```
204 No Content
```

---

## Business Rules

### Reservation Rules

#### Time Validation

- Cannot book past times
- End time must be after start time
- Bookings must be within school hours (6:00 AM - 8:00 PM)
- Maximum duration:
  - **Students**: 4 hours
  - **Staff**: Unlimited
- Date range queries cannot exceed 60 days

---

#### Overlap Prevention

- System checks for conflicting reservations
- Returns **400 Bad Request** if slot is already booked

---

#### Authorization

- Users can only cancel **their own** reservations
- Staff members can cancel **any** reservation

---

#### Privacy

Booking details are visible to:
- The person who made the booking
- Staff members

Other users only see time slots marked as **"booked"** without personal details.

---

## Email Notifications 📧

The system automatically sends emails for:

- ✅ **Booking Confirmation** – sent immediately after a successful reservation
- 🚫 **Cancellation Notice** – sent when a reservation is cancelled

### Gmail SMTP Setup

1. Enable **2-Factor Authentication** on your Google account
2. Generate an **App Password**  
   https://myaccount.google.com/apppasswords  
3. Select **Mail** and your device
4. Copy the generated 16-character password
5. Use this password as:

```bash
SMTP_PASSWORD=your-gmail-app-password
```

---

## Rate Limiting 🛡️

The API implements rate limiting to prevent abuse:

### OAuth Endpoints
- **Rate**: 5 requests per 12 seconds per IP
- **Applies to**:
  - `/oauth/login`
  - `/oauth/callback`

### API Endpoints
- **Rate**: 30 requests per 2 seconds per IP
- **Applies to**:
  - `/api/v1/reservations` (all methods)

When rate limit is exceeded, the API returns:
- **Status Code**: `429 Too Many Requests`
- **Retry-After**: Header indicating when to retry

---
