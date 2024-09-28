package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

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
		printAppointment(lastAppointment)
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
func GetMonthAppointments(db *gorm.DB, year int, month time.Month) ([]Appointment, error) {
	location, err := time.LoadLocation("Asia/Nicosia")
	if err != nil {
		panic(err)
	}
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, location)
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

func printAppointments(appointments []Appointment) {
	fmt.Printf("---------------------------------------------------\n")
	for _, appointment := range appointments {
		fmt.Printf("Appointment ID: %d, Start Time: %s\n", appointment.ID, appointment.StartTime.Format("15:04"))
		fmt.Printf("Appointment Duration: %d minutes, End Time: %s\n", appointment.Duration, appointment.StartTime.Add(time.Duration(appointment.Duration*int(time.Minute))).Format("15:04"))
		fmt.Printf("Patient Name: %s\n", appointment.Patient.Name)
		fmt.Printf("Appointment type: %s\n", appointment.AppointmentType.Description)
	}
	fmt.Printf("---------------------------------------------------\n\n")
}

func printAppointment(appointment Appointment) {
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

// CreatePatient creates a new patient record in the database
func CreatePatient(ctx context.Context, db *gorm.DB, patient Patient) (Patient, error) {
	if err := db.Create(&patient).Error; err != nil {
		return Patient{}, err
	}
	return patient, nil
}

// PrintPatient prints the patient details
func PrintPatient(patient Patient) {
	patient_json, _ := json.MarshalIndent(patient, "", "  ")
	fmt.Println(string(patient_json))
}
