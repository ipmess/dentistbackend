package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var db *gorm.DB
var ctx context.Context

func main() {
	// Load Database Endpoint configuration from local config.json file:
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("error loading configuration: %s\n", err)
		return
	}

	// Initialize the database connection with the custom endpoint
	db, err = initDB(config.DBEndpoint, config.Database, config.Username, config.Password)
	if err != nil {
		log.Fatalf("error initializing database at %s.\nFailed with '%s'\n", config.DBEndpoint, err)
		return
	}
	db.AutoMigrate(&Appointment{}, &Patient{}, &AppointmentType{})
	// Create context
	ctx = context.Background()

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
		PrintAppointment(createdAppointment)
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
	PrintAppointments(appointments)
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
	PrintAppointments(appointments)

	// Example usage of GetDayAppointments
	appointmentDay = time.Date(2024, 10, 25, 0, 0, 0, 0, time.Local)
	appointments, err = GetDayAppointments(db, appointmentDay)
	if err != nil {
		log.Fatalf("Error retrieving appointments: %s", err)
		return
	}

	// Print the appointments for the day
	fmt.Printf("Appointments for day %s :\n", appointmentDay)
	PrintAppointments(appointments)
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

	// implement a simple home page for the REST API:
	// Start the gorilla/mux server:
	router := mux.NewRouter()

	home := homeHandler{}

	// Register the routes
	router.HandleFunc("/", home.ServeHTTP)
	router.HandleFunc("/patients", NewPatient).Methods("POST")
	router.HandleFunc("/patients", ListPatients).Methods("GET")
	router.HandleFunc("/patients/{uuid}", GetPatient).Methods("GET")
	router.HandleFunc("/patients/{uuid}", UpdatePatient).Methods("PUT")
	router.HandleFunc("/patients/{uuid}", DeletePatient).Methods("DELETE")

	// Start the server
	http.ListenAndServe(":8080", router)
}
