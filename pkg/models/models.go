package models

import (
	"time"

	"gorm.io/gorm"
)

type Patient struct {
	// A structure to hold patient data
	gorm.Model
	ID                uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID              string        `gorm:"type:uuid;default:UUID();unique;not null" json:"UUID"`
	Name              string        `gorm:"type:varchar(255);not null" json:"Name"`
	PhoneNumber       string        `gorm:"type:varchar(20)" json:"PhoneNumber"`
	Email             string        `gorm:"type:varchar(255)" json:"Email"`
	Viber             bool          `json:"Viber"`
	Whatsapp          bool          `json:"Whatsapp"`
	SMS               bool          `json:"SMS"`
	EmailNotification bool          `json:"EmailNotification"`
	ReminderDays      int           `json:"ReminderDays"`
	Appointments      []Appointment `gorm:"foreignKey:PatientID"` // Relationship with Appointments
}

type Appointment struct {
	// A structure to hold appointment data
	gorm.Model
	ID                uint      `gorm:"primaryKey;autoIncrement"`
	UUID              string    `gorm:"type:uuid;default:UUID();unique;not null"`
	PatientID         uint      `gorm:"not null"` // Foreign key to Patients
	AppointmentTypeID uint      `gorm:"not null"` // Foreign key to AppointmentType
	StartTime         time.Time `gorm:"not null"` // Start time of the appointment
	Duration          int       `gorm:"not null"` // Duration in minutes
	Viber             bool
	Whatsapp          bool
	SMS               bool
	EmailNotification bool
	Reminder          int // Reminder in hours before the appointment
	// Relationships
	Patient         Patient         `gorm:"foreignKey:PatientID"`         // Belongs to Patient
	AppointmentType AppointmentType `gorm:"foreignKey:AppointmentTypeID"` // Belongs to AppointmentType
}

// AppointmentType represents a type of appointment in the system.
// It includes details such as a description, default duration, and color code.
// This structure is linked to the Appointment model through a foreign key relationship.
// There should only be a limited number of appointment types
type AppointmentType struct {
	gorm.Model
	ID              uint          `gorm:"primaryKey;autoIncrement"`
	Description     string        `gorm:"type:varchar(255);not null"`
	DefaultDuration int           `gorm:"not null"`                     // In minutes
	Color           string        `gorm:"type:char(7)"`                 // e.g. #FFA07A
	Appointments    []Appointment `gorm:"foreignKey:AppointmentTypeID"` // Relationship with Appointments
}
