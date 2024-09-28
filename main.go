package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Appointment struct {
	// A structure to hold appointment data
	gorm.Model
	ID                uint      `gorm:"primaryKey;autoIncrement"`
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

type Patient struct {
	// A structure to hold patient data
	gorm.Model
	ID                uint   `gorm:"primaryKey;autoIncrement"`
	Name              string `gorm:"type:varchar(255);not null"`
	PhoneNumber       string `gorm:"type:varchar(20)"`
	Email             string `gorm:"type:varchar(255)"`
	Viber             bool
	Whatsapp          bool
	SMS               bool
	EmailNotification bool
	ReminderDays      int
	Appointments      []Appointment `gorm:"foreignKey:PatientID"` // Relationship with Appointments
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

func main() {
	// Load Database Endpoint configuration from local config.json file:
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("error loading configuration: %s\n", err)
		return
	}

	// Initialize the database connection with the custom endpoint
	db, err := initDB(config.DBEndpoint, config.Database, config.Username, config.Password)
	if err != nil {
		log.Fatalf("error initializing database at %s.\nFailed with '%s'\n", config.DBEndpoint, err)
		return
	}
	db.AutoMigrate(&Appointment{}, &Patient{}, &AppointmentType{})
	// Create context
	ctx := context.Background()

	if config.PopulateDB {
		// Populate the database with sample data:
		fmt.Printf("Populating the database with sample data...\n")
		err = populate(ctx, db)
		if err != nil {
			log.Printf("couldn't populate database:\n %s\n", err)
		}
	}

	// Sample appointment data
	appointment := Appointment{
		PatientID:         1,
		AppointmentTypeID: 1,
		StartTime:         time.Now().AddDate(0, 1, 0),
		Duration:          60,
		Viber:             true,
		Whatsapp:          false,
		SMS:               true,
		EmailNotification: true,
		Reminder:          24,
	}

	// Call CreateAppointment with the db instance and sample data:
	createdAppointment, err := CreateAppointment(ctx, db, appointment)
	if err != nil {
		log.Printf("Error creating appointment: %s", err)
	} else {
		printAppointment(createdAppointment)
	}
	// Example usage of GetDayAppointments
	appointmentDay := time.Now().AddDate(0, 0, 20)
	appointments, err := GetDayAppointments(db, appointmentDay)
	if err != nil {
		log.Fatalf("Error retrieving appointments: %s", err)
		return
	}

	// Print the appointments for the day
	fmt.Printf("Appointments for day %s :\n", appointmentDay)
	printAppointments(appointments)
	fmt.Printf("End of list of appointments for day %s \n", appointmentDay)

	// Example usage of GetMonthAppointments

	year := time.Now().Year()
	month := time.Now().Month()

	appointments, err = GetMonthAppointments(db, year, month)
	if err != nil {
		log.Fatalf("Error retrieving appointments:\n%s", err)
		return
	}
	// Print the Month's appointments
	fmt.Printf("Appointments for month %s :\n", time.Now().Format("January 2006"))
	printAppointments(appointments)

	// Example usage of GetDayAppointments
	appointmentDay = time.Date(2024, 10, 25, 0, 0, 0, 0, time.Local)
	appointments, err = GetDayAppointments(db, appointmentDay)
	if err != nil {
		log.Fatalf("Error retrieving appointments: %s", err)
		return
	}

	// Print the appointments for the day
	fmt.Printf("Appointments for day %s :\n", appointmentDay)
	printAppointments(appointments)
	fmt.Printf("End of list of appointments for day %s \n", appointmentDay)

	// sample Patient data:
	patient := Patient{
		Name:              "Ανδρέας Ανδρέου",
		PhoneNumber:       "99798979",
		Email:             "andreou@example.com",
		Viber:             true,
		Whatsapp:          false,
		SMS:               true,
		EmailNotification: true,
		ReminderDays:      100,
	}

	pat, err := CreatePatient(ctx, db, patient)
	if err != nil {
		log.Printf("Error creating patient: %s", patient.Name)
		log.Fatalf("Error creating patient:\n%s", err)
		return
	}

	fmt.Printf("Patient created successfully: %s\n", pat.Name)
	PrintPatient(pat)
}
