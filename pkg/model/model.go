package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aquareum.tv/aquareum/pkg/log"
	"github.com/lmittmann/tint"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBModel struct {
	DB *gorm.DB
}

type Model interface {
	CreateNotification(token string) error
	ListNotifications() ([]Notification, error)

	CreatePlayerEvent(event PlayerEventAPI) error
}

type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func MakeDB(dbURL string) (Model, error) {
	log.Log(context.Background(), "starting database", "dbURL", dbURL)
	if !strings.HasPrefix(dbURL, "sqlite://") {
		dbURL = fmt.Sprintf("sqlite://%s", dbURL)
	}
	sqliteSuffix := dbURL[len("sqlite://"):]
	// if this isn't ":memory:", ensure that directory exists (eg, if db
	// file is being initialized)
	if !strings.Contains(sqliteSuffix, ":?") {
		os.MkdirAll(filepath.Dir(sqliteSuffix), os.ModePerm)
	}
	dial := sqlite.Open(sqliteSuffix)

	gormLogger := slogGorm.New(slogGorm.WithHandler(tint.NewHandler(os.Stderr, &tint.Options{
		TimeFormat: time.RFC3339,
	})))

	db, err := gorm.Open(dial, &gorm.Config{
		SkipDefaultTransaction: true,
		TranslateError:         true,
		Logger:                 gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("error starting database: %w", err)
	}
	for _, model := range []any{Notification{}, PlayerEvent{}} {
		err = db.AutoMigrate(model)
		if err != nil {
			return nil, err
		}
	}
	return &DBModel{DB: db}, nil
}
