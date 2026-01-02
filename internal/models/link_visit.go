package models

import "time"

// LinkVisit представляет запись о посещении ссылки
type LinkVisit struct {
	ID        int32     `json:"id"`
	LinkID    int32     `json:"link_id"`
	IP        string    `json:"ip"`
	UserAgent *string   `json:"user_agent,omitempty"`
	Referer   *string   `json:"referer,omitempty"`
	Status    int32     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
