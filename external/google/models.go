package google

import "time"

type Event struct {
    ID string `json:"id"`
}

type AccessToken struct {
    Token     string    `json:"access_token"`
    ExpiresIn int       `json:"expires_in"`
    CreatedAt time.Time `json:"-"`
}

func (t *AccessToken) IsExpired() bool {
    expiryTime := t.CreatedAt.Add(time.Duration(t.ExpiresIn-60) * time.Second)
    return time.Now().After(expiryTime)
}

type EventRequest struct {
    Summary     string         `json:"summary"`
    Description string         `json:"description"`
    Start       DateTimeObject `json:"start"`
    End         DateTimeObject `json:"end"`
}

type DateTimeObject struct {
    DateTime string `json:"dateTime"`
}