package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Activity struct {
	ID        string `gorm:"primaryKey;type:char(36)"`
	Name      string `gorm:"type:varchar(255);not null"`
	ColorCode string `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time
}

func (a *Activity) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return
}

type ActivityLog struct {
	ID         string     `gorm:"primaryKey;type:char(36)"`
	ActivityID string     `gorm:"type:char(36);index;not null"`
	StartTime  time.Time
	EndTime    *time.Time
	CreatedAt  time.Time
	Memo       string `gorm:"type:text"`
}

func (l *ActivityLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return
}

type DailyPlan struct {
	ID             string    `gorm:"primaryKey;type:char(36)"`
	ActivityID     string    `gorm:"type:char(36);index;not null"`
	Date           string    `gorm:"type:varchar(10);index;not null"` // YYYY-MM-DD
	StartMinute    int       `gorm:"not null;default:0"` // minutes from midnight
	PlannedMinutes int       `gorm:"not null;default:30"`
	CreatedAt      time.Time
	Memo           string    `gorm:"type:text"`
}

func (p *DailyPlan) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}
