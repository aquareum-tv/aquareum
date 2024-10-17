package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PlayerEventAPI struct {
	ID        string         `json:"id"`
	Time      time.Time      `json:"time"`
	PlayerId  string         `json:"playerId"`
	EventType string         `json:"eventType"`
	Meta      map[string]any `json:"meta"`
}

type PlayerEvent struct {
	ID        string `gorm:"primarykey"`
	Time      time.Time
	PlayerId  string `gorm:"index"`
	EventType string
	Meta      JSON
}

func MaybeNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{Valid: true, String: s}
}

func (m *DBModel) CreatePlayerEvent(event PlayerEventAPI) error {
	uu, err := uuid.NewV7()
	if err != nil {
		return err
	}
	err = m.DB.Model(PlayerEvent{}).Create(PlayerEvent{
		ID:        uu.String(),
		Time:      event.Time,
		PlayerId:  event.PlayerId,
		EventType: event.EventType,
	}).Error
	if err != nil {
		return err
	}
	return nil
}
