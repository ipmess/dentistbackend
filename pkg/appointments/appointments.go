package appointments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ipmess/dentistbackend/pkg/models"
	"gorm.io/gorm"
)

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

type HTTPHandler struct {
	DB  *gorm.DB
	Ctx context.Context
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

func CreateAppointment(ctx context.Context, db *gorm.DB, appointment models.Appointment) (models.Appointment, error) {
	// Use the passed db descriptor to create the appointment record
	db.Model(&appointment).Association("Patient")
	db.Model(&appointment).Association("AppointmentType")

	// sanity check: Verify that the appointment's patient and appointment type exist in the database:
	var patient models.Patient
	err := db.First(&patient, appointment.PatientID).Error
	if err != nil {
		log.Fatalf("Error retrieving patient with ID %d.\nFailed with '%s'\n", appointment.PatientID, err)
	}
	var appointmentType models.AppointmentType
	err = db.First(&appointmentType, appointment.AppointmentTypeID).Error
	if err != nil {
		log.Fatalf("Error retrieving appointment type with ID %d.\nFailed with '%s'\n", appointment.AppointmentTypeID, err)
	}
	// store retrieved patient data and appointment type into the appointment struct:
	appointment.Patient = patient
	appointment.AppointmentType = appointmentType

	// Check if the appointment overlaps with an existing appointment:

	var lastAppointment models.Appointment
	//startTime := appointment.StartTime
	endTime := appointment.StartTime.Add(time.Minute * time.Duration(appointment.Duration))
	fmt.Printf("Checking for overlapping appointments for appointments between %s and %s\n", appointment.StartTime.Format("02 Jan 2006 15:04 -0700"), appointment.StartTime.Add(time.Minute*time.Duration(appointment.Duration)).Format("02 Jan 2006 15:04 -0700"))

	// Find the last appointment before the new appointment's end time
	SQLStatement := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&models.Appointment{}).Where("start_time < ?", endTime).Order("start_time desc").First(&lastAppointment)
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
		return models.Appointment{}, err
	}
	err = db.Model(&appointment).Association("Patient").Error
	if err != nil {
		log.Fatalf(
			"association error (of appointment with patient) when creating appointment.\nFailed with '%s'\n",
			err,
		)
		return models.Appointment{}, err
	}
	err = db.Model(&appointment).Association("AppointmentType").Error
	if err != nil {
		log.Fatalf(
			"association error (of appointment with AppointmentType) when creating appointment.\nFailed with '%s'\n",
			err,
		)
		return models.Appointment{}, err
	}
	return appointment, nil
}

// GetMonthAppointments retrieves all appointments for a specific month
func GetMonthAppointments(db *gorm.DB, appointmentDate time.Time) ([]models.Appointment, error) {
	location, err := time.LoadLocation("Asia/Nicosia")
	if err != nil {
		panic(err)
	}
	startDate := time.Date(appointmentDate.Year(), appointmentDate.Month(), 1, 0, 0, 0, 0, location)
	endDate := startDate.AddDate(0, 1, -1).Add(24 * time.Hour) // Last day of the month + 1 day
	var appointments []models.Appointment
	err = db.Preload("Patient").Preload("AppointmentType").Where("start_time BETWEEN ? AND ?", startDate, endDate).Order("start_time asc").Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

// GetDayAppointments retrieves all appointments for a specific day
func GetDayAppointments(db *gorm.DB, date time.Time) ([]models.Appointment, error) {

	// Normalize the date to the start of the day
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	// End of the day (23:59:59.999)
	endDate := startDate.Add(24 * time.Hour).Add(-time.Millisecond)
	var appointments []models.Appointment

	err := db.Preload("Patient").Preload("AppointmentType").Where("start_time BETWEEN ? AND ?", startDate, endDate).Order("start_time asc").Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func PrintAppointments(appointments []models.Appointment) {
	fmt.Printf("---------------------------------------------------\n")
	for _, appointment := range appointments {
		fmt.Printf("Appointment ID: %d, Start Time: %s\n", appointment.ID, appointment.StartTime.Format("15:04"))
		fmt.Printf("Appointment Duration: %d minutes, End Time: %s\n", appointment.Duration, appointment.StartTime.Add(time.Duration(appointment.Duration*int(time.Minute))).Format("15:04"))
		fmt.Printf("Patient Name: %s\n", appointment.Patient.Name)
		fmt.Printf("Appointment type: %s\n", appointment.AppointmentType.Description)
	}
	fmt.Printf("---------------------------------------------------\n\n")
}

func PrintAppointment(appointment models.Appointment) {
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
		Patient           models.Patient
		AppointmentType   models.AppointmentType
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

func (h *HTTPHandler) NewAppointment(w http.ResponseWriter, r *http.Request) {
	var appointment models.Appointment
	err := json.NewDecoder(r.Body).Decode(&appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()
	appointment.ID = 0
	tempUUID, _ := uuid.NewV7()
	appointment.UUID = tempUUID.String()

	appointment, err = CreateAppointment(h.Ctx, h.DB, appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointment)
}

func (h *HTTPHandler) ListAppointments(w http.ResponseWriter, r *http.Request) {
	var request appointmentRequest
	//  check whether the request is empty:
	if r.Body == nil {
		request.Frame = "day"
		request.StartDate = time.Now().Format("02-01-2006")
	} else {
		// the request should include a time frame and a start date:
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := request.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appointments []models.Appointment

	// convert request.StartDate from string to time.Time:
	requestStartTime, err := time.Parse("02-01-2006", request.StartDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch request.Frame {
	case Day:
		appointments, err = GetDayAppointments(h.DB, requestStartTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Month:
		appointments, err = GetMonthAppointments(h.DB, requestStartTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Week:
		monthsAppointments, err := GetMonthAppointments(h.DB, requestStartTime)
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

			monthsAppointments, err := GetMonthAppointments(h.DB, requestStartTime.AddDate(0, m, 0))
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

func (h *HTTPHandler) GetAppointment(w http.ResponseWriter, r *http.Request) {
	// Get the appointment UUID from the URL:
	vars := mux.Vars(r)
	appointmentUUID := vars["uuid"]

	var appointment models.Appointment
	// Find the appointment with the given UUID:
	err := h.DB.Preload("Patient").Preload("AppointmentType").Where("uuid = ?", appointmentUUID).First(&appointment).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointment)
}

func (h *HTTPHandler) UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	// Get the appointment UUID from the URL:
	vars := mux.Vars(r)
	appointmentUUID := vars["uuid"]

	var newAppointment models.Appointment
	err := json.NewDecoder(r.Body).Decode(&newAppointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appointmentInDB models.Appointment
	// Find the appointment with the given UUID:
	err = h.DB.Where("uuid = ?", appointmentUUID).First(&appointmentInDB).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// TODO:
	// Refactor the UpdateAppointment function to use the .Update GORM method
	// instead of creating a new appointment and saving it to the database.

	// Update the appointment:
	err = h.DB.Save(&newAppointment).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appointmentInDB.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newAppointment)
}

func (h *HTTPHandler) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	// Get the appointment UUID from the URL:
	vars := mux.Vars(r)
	appointmentUUID := vars["uuid"]

	var appointment models.Appointment
	// Find the appointment with the given UUID:
	err := h.DB.Where("uuid = ?", appointmentUUID).First(&appointment).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Delete the appointment:
	err = h.DB.Delete(&appointment).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
