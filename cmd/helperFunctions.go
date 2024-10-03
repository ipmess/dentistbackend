package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	DBEndpoint string `json:"db_endpoint"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
	PopulateDB bool   `json:"populate_db"`
}

func initDB(endpoint, database, username, password string) (*gorm.DB, error) {
	dsn := username + ":" + password + "@tcp(" + endpoint + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
	// Open the database connection:
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database at %s: %w", endpoint, err)
	}
	return db, nil
}

func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(bytes, &config)
	return config, err
}

func populate(ctx context.Context, db *gorm.DB) error {
	// Load sample data from JSON files
	// Start by loading the sample appointment types:
	appointmentTypesFilename := "apptypes.json"
	var AppointmentTypes []AppointmentType
	file, err := os.Open(appointmentTypesFilename)
	if err != nil {
		log.Printf("Error opening file %s: %s\n", appointmentTypesFilename, err)
		return err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file %s: %s\n", appointmentTypesFilename, err)
		return err
	}

	err = json.Unmarshal(bytes, &AppointmentTypes)
	if err != nil {
		log.Printf("Error unmarshalling JSON from file %s: %s\n", appointmentTypesFilename, err)
		return err
	}
	// write sample appointment data into the AppointmentType table:
	result := db.Create(&AppointmentTypes)
	if result.Error != nil {
		log.Printf("Error creating appointment types:\n %s\n", result.Error)
		log.Printf("---------------------------------------------------\n")
		for _, apt := range AppointmentTypes {
			log.Printf("Appointment type ID: %d, Description: %s\n", apt.ID, apt.Description)
			log.Printf("Appointment DefaultDuration: %d minutes, Color: %s\n", apt.DefaultDuration, apt.Color)
		}
		log.Printf("---------------------------------------------------\n\n")
		return result.Error
	}
	fmt.Println("Appointment types created successfully")

	// next, we will write sample patient data into the Patient table:
	patientsFilename := "patients-utf8.json"
	var samplePatients []Patient
	file, err = os.Open(patientsFilename)
	if err != nil {
		log.Printf("Error opening file %s: %s\n", patientsFilename, err)
		return err
	}
	defer file.Close()
	bytes, err = io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file %s: %s\n", patientsFilename, err)
		return err
	}

	err = json.Unmarshal(bytes, &samplePatients)
	if err != nil {
		log.Printf("Error unmarshalling JSON from sample patient data file %s:\n %s\n", patientsFilename, err)
		return err
	}
	// write sample patient data into the Patient table:
	result = db.Create(&samplePatients)
	if result.Error != nil {
		log.Printf("Error creating patients:\n %s\n", result.Error)
		return result.Error
	}
	fmt.Println("Patients created successfully")

	// Finally, we will write sample appointment data into the Appointment table:
	// Only this time we have to create 100 appointments spread out throughout the next 2 months:

	// read all appointment types from the database:
	var appointmentTypes []AppointmentType
	result = db.Find(&appointmentTypes)
	if result.Error != nil {
		log.Printf("Error retrieving appointment types from the database:\n %s\n", result.Error)
		return result.Error
	}

	// read all patients from the database:
	var patients []Patient
	result = db.Find(&patients)
	if result.Error != nil {
		log.Printf("Error retrieving patients from the database:\n %s\n", result.Error)
		return result.Error
	}

	// First off, we need to add a single appointment so that the appointment overlap check does not fail due to an empty appointments table:
	appointmentType := appointmentTypes[rand.Intn(len(appointmentTypes))]
	patient := patients[rand.Intn(len(patients))]
	startTime := time.Now()
	appointment := Appointment{
		PatientID:         patient.ID,
		AppointmentTypeID: appointmentType.ID,
		StartTime:         startTime,
		Duration:          appointmentType.DefaultDuration,
		Viber:             patient.Viber,
		Whatsapp:          patient.Whatsapp,
		SMS:               patient.SMS,
		EmailNotification: patient.EmailNotification,
		Reminder:          rand.Intn(48) + 12,
		Patient:           patient,
		AppointmentType:   appointmentType,
	}

	if err := db.Create(&appointment).Error; err != nil {
		log.Printf("Error creating initial appointment:\n %s\n", err)
		return err
	}

	// We will create 100 appointments:
	for i := 0; i < 100; i++ {
		// Randomly select an appointment type from appointmentTypes:
		appointmentType = appointmentTypes[rand.Intn(len(appointmentTypes))]
		// Randomly select a patient from patients:
		patient = patients[rand.Intn(len(patients))]
		// Randomly select a date and time within the next 2 months.
		// The hours have to be between 8 and 17, and the minutes have to be either 0 or 30:
		// The days have to be working days (Monday to Friday):
		// The months have to be the current month or the next month:
		// The year has to be the current year:
		// The duration has to be the appointment type's DefaultDuration:
		// The notification preferences have to be the patient's preferences:
		// The reminder time has to be between 1 and 48 hours before the appointment:
		// Create the appointment:
		startTime = time.Now().AddDate(0, rand.Intn(2), rand.Intn(30))
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 8+rand.Intn(10), rand.Intn(2)*30, 0, 0, startTime.Location())
		//make sure the date is a working day:
		for startTime.Weekday() == time.Saturday || startTime.Weekday() == time.Sunday {
			startTime = startTime.AddDate(0, 0, 1)
		}
		appointment = Appointment{
			PatientID:         patient.ID,
			AppointmentTypeID: appointmentType.ID,
			StartTime:         startTime,
			Duration:          appointmentType.DefaultDuration,
			Viber:             patient.Viber,
			Whatsapp:          patient.Whatsapp,
			SMS:               patient.SMS,
			EmailNotification: patient.EmailNotification,
			Reminder:          rand.Intn(48) + 12,
			Patient:           patient,
			AppointmentType:   appointmentType,
		}

		_, err = CreateAppointment(ctx, db, appointment)
		if err != nil {
			log.Printf("Couldn't create random appointment: %s\n", err)
		}

	}
	return nil
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Appointment Scheduler API"))
}
