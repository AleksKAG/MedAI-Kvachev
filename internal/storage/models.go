package storage

import "gorm.io/gorm"

type LabResult struct {
	ID        uint    `gorm:"primaryKey"`
	PatientID string
	TestName  string // "hemoglobin", "glucose"
	Value     float64
	Unit      string
	Date      int64  // Unix timestamp
}

type Symptom struct {
	ID        uint   `gorm:"primaryKey"`
	PatientID string
	Name      string // "fever", "weakness"
	Date      int64
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&LabResult{}, &Symptom{})
}
