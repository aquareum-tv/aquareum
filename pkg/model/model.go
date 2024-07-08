package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aquareum.tv/aquareum/pkg/log"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Notification struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Token     string
}

func MakeDB(dbURL string) (*gorm.DB, error) {
	log.Log(context.Background(), "starting database", "dbURL", dbURL)
	if !strings.HasPrefix(dbURL, "sqlite://") {
		return nil, fmt.Errorf("only sqlite:// urls currently supported, got %s", dbURL)
	}
	sqliteSuffix := dbURL[len("sqlite://"):]
	// if this isn't ":memory:", ensure that directory exists (eg, if db
	// file is being initialized)
	if !strings.Contains(sqliteSuffix, ":?") {
		os.MkdirAll(filepath.Dir(sqliteSuffix), os.ModePerm)
	}
	dial := sqlite.Open(sqliteSuffix)

	gormLogger := slogGorm.New()

	db, err := gorm.Open(dial, &gorm.Config{
		SkipDefaultTransaction: true,
		TranslateError:         true,
		Logger:                 gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("error starting database: %w", err)
	}
	err = db.AutoMigrate(Notification{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
