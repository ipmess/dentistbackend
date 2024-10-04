package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ipmess/dentistbackend/pkg/appointments"
	"github.com/ipmess/dentistbackend/pkg/authenticationHelper"
	"github.com/ipmess/dentistbackend/pkg/models"
	"github.com/ipmess/dentistbackend/pkg/patient"
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
	db.AutoMigrate(&models.Appointment{}, &models.Patient{}, &models.AppointmentType{})

	// Create context
	ctx = context.Background()

	patientHandler := patient.HTTPHandler{
		DB:  db,
		Ctx: ctx,
	}

	appointmentHandler := appointments.HTTPHandler{
		DB:  db,
		Ctx: ctx,
	}

	if config.PopulateDB {
		// Populate the database with sample data:
		fmt.Printf("Populating the database with sample data...\n")
		err = populate(ctx, db)
		if err != nil {
			log.Printf("couldn't populate database:\n %s\n", err)
		}
	}

	// Sample appointment data
	sampleAppointment := models.Appointment{
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
	createdAppointment, err := appointments.CreateAppointment(ctx, db, sampleAppointment)
	if err != nil {
		log.Printf("Error creating appointment: %s", err)
	} else {
		appointments.PrintAppointment(createdAppointment)
	}

	// Example usage of GetDayAppointments
	appointmentDay := time.Now().AddDate(0, 0, 20)
	dayAppointments, err := appointments.GetDayAppointments(db, appointmentDay)
	if err != nil {
		log.Fatalf("Error retrieving appointments: %s", err)
		return
	}

	// Print the appointments for the day
	fmt.Printf("Appointments for day %s :\n", appointmentDay)
	appointments.PrintAppointments(dayAppointments)
	fmt.Printf("End of list of appointments for day %s \n", appointmentDay)

	// Example usage of GetMonthAppointments

	monthAppointments, err := appointments.GetMonthAppointments(db, time.Now())
	if err != nil {
		log.Fatalf("Error retrieving appointments:\n%s", err)
		return
	}
	// Print the Month's appointments
	fmt.Printf("Appointments for month %s :\n", time.Now().Format("January 2006"))
	appointments.PrintAppointments(monthAppointments)

	// Example usage of GetDayAppointments
	appointmentDay = time.Date(2024, 10, 25, 0, 0, 0, 0, time.Local)
	dayAppointments, err = appointments.GetDayAppointments(db, appointmentDay)
	if err != nil {
		log.Fatalf("Error retrieving appointments: %s", err)
		return
	}

	// Print the appointments for the day
	fmt.Printf("Appointments for day %s :\n", appointmentDay)
	appointments.PrintAppointments(dayAppointments)
	fmt.Printf("End of list of appointments for day %s \n", appointmentDay)

	// sample Patient data:
	samplePatient := models.Patient{
		Name:              "Ανδρέας Ανδρέου",
		PhoneNumber:       "99798979",
		Email:             "andreou@example.com",
		Viber:             true,
		Whatsapp:          false,
		SMS:               true,
		EmailNotification: true,
		ReminderDays:      100,
	}

	pat, err := patient.CreatePatient(ctx, db, samplePatient)
	if err != nil {
		log.Printf("Error creating patient: %s", samplePatient.Name)
		log.Fatalf("Error creating patient:\n%s", err)
		return
	}

	fmt.Printf("Patient created successfully: %s\n", pat.Name)
	patient.PrintPatient(pat)

	// implement a simple home page for the REST API:
	// Start the gorilla/mux server:
	router := mux.NewRouter()

	// Register the routes
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/patients", patientHandler.NewPatient).Methods("POST")
	router.HandleFunc("/patients", patientHandler.ListPatients).Methods("GET")
	router.HandleFunc("/patients/{uuid}", patientHandler.GetPatient).Methods("GET")
	router.HandleFunc("/patients/{uuid}", patientHandler.UpdatePatient).Methods("PUT")
	router.HandleFunc("/patients/{uuid}", patientHandler.DeletePatient).Methods("DELETE")
	router.HandleFunc("/appointments", appointmentHandler.NewAppointment).Methods("POST")
	router.HandleFunc("/appointments/date", appointmentHandler.ListAppointments).Methods("GET")
	router.HandleFunc("/appointments/{uuid}", appointmentHandler.GetAppointment).Methods("GET")
	router.HandleFunc("/appointments/{uuid}", appointmentHandler.UpdateAppointment).Methods("PUT")
	router.HandleFunc("/appointments/{uuid}", appointmentHandler.DeleteAppointment).Methods("DELETE")
	router.HandleFunc("/authenticate", authenticationHelper.Authenticate).Methods("GET")

	// Start the server
	http.ListenAndServe(":8080", router)
}

/*
* `POST /appointments` to create an appointment
* `GET /appointments/month` to get a list of all appointments for a particular month/year.
* `GET /appointments/week` to get a list of all appointments for a particular week/year.
* `GET /appointments/date` to get a list of all appointments for a particular date.
* `GET /appointments/:uuid` to get a specific appointment
* `PUT /appointments/:uuid` to update a specific appointment
* `DELETE /appointments/:uuid` to delete a specific appointment
 */
