package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Appointment struct {
	// A structure to hold appointment data
	gorm.Model
	ID                uint      `gorm:"primaryKey;autoIncrement"`
	uuid              string    `gorm:"type:uuid;default:UUIDv7();unique;not null"`
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

// limit the appointmentRequest.timeFrame to the following values:
// day, week, month, year
type TimeFrame string

const (
	Day   TimeFrame = "day"
	Week  TimeFrame = "week"
	Month TimeFrame = "month"
	Year  TimeFrame = "year"
)

type appointmentRequest struct {
	// a structure to hold the request data for listing appointments
	Frame     TimeFrame
	StartDate string
}

// Validate checks if the timeFrame is one of the allowed values
func (ar *appointmentRequest) Validate() error {
	switch ar.Frame {
	case Day, Week, Month, Year:
		return nil
	default:
		return errors.New("invalid time frame")
	}
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

func CreateAppointment(ctx context.Context, db *gorm.DB, appointment Appointment) (Appointment, error) {
	// Use the passed db descriptor to create the appointment record
	db.Model(&appointment).Association("Patient")
	db.Model(&appointment).Association("AppointmentType")

	// sanity check: Verify that the appointment's patient and appointment type exist in the database:
	var patient Patient
	err := db.First(&patient, appointment.PatientID).Error
	if err != nil {
		log.Fatalf("Error retrieving patient with ID %d.\nFailed with '%s'\n", appointment.PatientID, err)
	}
	var appointmentType AppointmentType
	err = db.First(&appointmentType, appointment.AppointmentTypeID).Error
	if err != nil {
		log.Fatalf("Error retrieving appointment type with ID %d.\nFailed with '%s'\n", appointment.AppointmentTypeID, err)
	}
	// store retrieved patient data and appointment type into the appointment struct:
	appointment.Patient = patient
	appointment.AppointmentType = appointmentType

	// Check if the appointment overlaps with an existing appointment:

	var lastAppointment Appointment
	//startTime := appointment.StartTime
	endTime := appointment.StartTime.Add(time.Minute * time.Duration(appointment.Duration))
	fmt.Printf("Checking for overlapping appointments for appointments between %s and %s\n", appointment.StartTime.Format("02 Jan 2006 15:04 -0700"), appointment.StartTime.Add(time.Minute*time.Duration(appointment.Duration)).Format("02 Jan 2006 15:04 -0700"))

	// Find the last appointment before the new appointment's end time
	SQLStatement := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&Appointment{}).Where("start_time < ?", endTime).Order("start_time desc").First(&lastAppointment)
	})
	fmt.Printf("SQL statement:\n%s\n", SQLStatement)
	err = db.Preload("Patient").Preload("AppointmentType").Where("start_time < ?", endTime).Order("start_time desc").First(&lastAppointment).Error
	if err != nil {
		log.Fatalf("Error checking for overlapping appointments.\nFailed with '%s'\n", err)
	}

	// Check if the new appointment's start time is after the last appointment's end time:

	lastAppointmentsEndTime := lastAppointment.StartTime.Add(time.Duration(lastAppointment.Duration) * time.Minute)
	if lastAppointmentsEndTime.After(appointment.StartTime) {
		fmt.Printf("Overlapping appointment:\n")
		PrintAppointment(lastAppointment)
		return appointment, fmt.Errorf("appointment overlaps with an existing appointment")
	}

	// Now we know it is safe to create the appointment:

	if err := db.Create(&appointment).Error; err != nil {
		return Appointment{}, err
	}
	err = db.Model(&appointment).Association("Patient").Error
	if err != nil {
		log.Fatalf(
			"association error (of appointment with patient) when creating appointment.\nFailed with '%s'\n",
			err,
		)
		return Appointment{}, err
	}
	err = db.Model(&appointment).Association("AppointmentType").Error
	if err != nil {
		log.Fatalf(
			"association error (of appointment with AppointmentType) when creating appointment.\nFailed with '%s'\n",
			err,
		)
		return Appointment{}, err
	}
	return appointment, nil
}

// GetMonthAppointments retrieves all appointments for a specific month
func GetMonthAppointments(db *gorm.DB, appointmentDate time.Time) ([]Appointment, error) {
	location, err := time.LoadLocation("Asia/Nicosia")
	if err != nil {
		panic(err)
	}
	startDate := time.Date(appointmentDate.Year(), appointmentDate.Month(), 1, 0, 0, 0, 0, location)
	endDate := startDate.AddDate(0, 1, -1).Add(24 * time.Hour) // Last day of the month + 1 day
	var appointments []Appointment
	err = db.Preload("Patient").Preload("AppointmentType").Where("start_time BETWEEN ? AND ?", startDate, endDate).Order("start_time asc").Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

// GetDayAppointments retrieves all appointments for a specific day
func GetDayAppointments(db *gorm.DB, date time.Time) ([]Appointment, error) {

	// Normalize the date to the start of the day
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	// End of the day (23:59:59.999)
	endDate := startDate.Add(24 * time.Hour).Add(-time.Millisecond)
	var appointments []Appointment

	err := db.Preload("Patient").Preload("AppointmentType").Where("start_time BETWEEN ? AND ?", startDate, endDate).Order("start_time asc").Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func PrintAppointments(appointments []Appointment) {
	fmt.Printf("---------------------------------------------------\n")
	for _, appointment := range appointments {
		fmt.Printf("Appointment ID: %d, Start Time: %s\n", appointment.ID, appointment.StartTime.Format("15:04"))
		fmt.Printf("Appointment Duration: %d minutes, End Time: %s\n", appointment.Duration, appointment.StartTime.Add(time.Duration(appointment.Duration*int(time.Minute))).Format("15:04"))
		fmt.Printf("Patient Name: %s\n", appointment.Patient.Name)
		fmt.Printf("Appointment type: %s\n", appointment.AppointmentType.Description)
	}
	fmt.Printf("---------------------------------------------------\n\n")
}

func PrintAppointment(appointment Appointment) {
	type appointmentForPrinting struct {
		// Define a struct for printing the appointment. Similar fields to the original appointment struct, except create a string for the start time:
		ID                uint
		PatientID         uint
		AppointmentTypeID uint
		StartTime         string
		Duration          int
		Viber             bool
		Whatsapp          bool
		SMS               bool
		EmailNotification bool
		Reminder          int
		Patient           Patient
		AppointmentType   AppointmentType
	}
	var appointmentForPrint appointmentForPrinting
	appointmentForPrint.ID = appointment.ID
	appointmentForPrint.PatientID = appointment.PatientID
	appointmentForPrint.AppointmentTypeID = appointment.AppointmentTypeID
	appointmentForPrint.StartTime = appointment.StartTime.Format("02 Jan 2006 15:04")
	appointmentForPrint.Duration = appointment.Duration
	appointmentForPrint.Viber = appointment.Viber
	appointmentForPrint.Whatsapp = appointment.Whatsapp
	appointmentForPrint.SMS = appointment.SMS
	appointmentForPrint.EmailNotification = appointment.EmailNotification
	appointmentForPrint.Reminder = appointment.Reminder
	appointmentForPrint.Patient = appointment.Patient
	appointmentForPrint.AppointmentType = appointment.AppointmentType

	// Marshal the appointmentForPrint struct into a JSON string:

	json_appointment, _ := json.MarshalIndent(appointmentForPrint, "", "  ")
	fmt.Println(string(json_appointment))
}

func NewAppointment(w http.ResponseWriter, r *http.Request) {
	var appointment Appointment
	err := json.NewDecoder(r.Body).Decode(&appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()
	appointment.ID = 0
	tempUUID, _ := uuid.NewV7()
	appointment.uuid = tempUUID.String()

	appointment, err = CreateAppointment(ctx, db, appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointment)
}

func ListAppointments(w http.ResponseWriter, r *http.Request) {
	// the request should include a time frame and a start date:
	var request appointmentRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = request.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appointments []Appointment

	// convert request.StartDate from string to time.Time:
	requestStartTime, err := time.Parse("02-01-2006", request.StartDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch request.Frame {
	case Day:
		appointments, err = GetDayAppointments(db, requestStartTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Month:
		appointments, err = GetMonthAppointments(db, requestStartTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Week:
		monthsAppointments, err := GetMonthAppointments(db, requestStartTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Filter the appointments to only include the appointments for the week:
		for _, appointment := range monthsAppointments {
			if appointment.StartTime.Before(requestStartTime.AddDate(0, 0, 7)) {
				appointments = append(appointments, appointment)
			}
		}

	case Year:

		for m := 0; m < 12; m++ {

			monthsAppointments, err := GetMonthAppointments(db, requestStartTime.AddDate(0, m, 0))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// append the Month's appointments to the appointments slice:
			appointments = append(appointments, monthsAppointments...)
		}

	default:
		http.Error(w, "invalid time frame", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointments)

}
