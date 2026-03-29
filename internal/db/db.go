package db

import (
	"log"

	"dailyApp/backend/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schemas
	err = db.AutoMigrate(&model.Activity{}, &model.ActivityLog{}, &model.DailyPlan{})
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return nil, err
	}

	return db, nil
}
