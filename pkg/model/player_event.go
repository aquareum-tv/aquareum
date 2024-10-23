package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
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
	Meta      datatypes.JSON
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
	metaBs, err := json.Marshal(event.Meta)
	if err != nil {
		return err
	}
	err = m.DB.Model(PlayerEvent{}).Create(PlayerEvent{
		ID:        uu.String(),
		Time:      event.Time,
		PlayerId:  event.PlayerId,
		EventType: event.EventType,
		Meta:      datatypes.JSON(metaBs),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *DBModel) ListPlayerEvents(playerId string) ([]PlayerEvent, error) {
	events := []PlayerEvent{}
	// err := m.DB.Find(&events).Error
	err := m.DB.Where("player_id = ?", playerId).Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("error retrieving player events: %w", err)
	}
	return events, nil
}

func (m *DBModel) PlayerReport(playerId string) (map[string]float64, error) {
	events, err := m.ListPlayerEvents(playerId)
	if err != nil {
		return nil, err
	}
	report := map[string]float64{}
	for _, e := range events {
		if e.EventType != "aq-played" {
			continue
		}
		bs, err := e.Meta.MarshalJSON()
		if err != nil {
			return nil, err
		}
		meta := map[string]any{}
		err = json.Unmarshal(bs, &meta)
		if err != nil {
			return nil, err
		}
		whatHappenedAny, ok := meta["whatHappened"]
		if !ok {
			continue
		}
		whatHappened, ok := whatHappenedAny.(map[string]any)
		if !ok {
			continue
		}
		for state, time := range whatHappened {
			ms, ok := time.(float64)
			if ok {
				report[state] = report[state] + ms
			}
		}
	}
	return report, nil
}

func (m *DBModel) ClearPlayerEvents() error {
	return m.DB.Where("1 = 1").Delete(&PlayerEvent{}).Error
}
