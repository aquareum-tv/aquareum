package model

import (
	"context"
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
}

type Notification struct {
	Token     string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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
	err = db.AutoMigrate(Notification{})
	if err != nil {
		return nil, err
	}
	return &DBModel{DB: db}, nil
}

func (m *DBModel) CreateNotification(token string) error {
	err := m.DB.Model(Notification{}).Create(&Notification{
		Token: token,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *DBModel) ListNotifications() ([]Notification, error) {
	nots := []Notification{}
	err := m.DB.Find(&nots).Error
	if err != nil {
		return nil, fmt.Errorf("error retrieving notifications: %w", err)
	}
	return nots, nil
}
