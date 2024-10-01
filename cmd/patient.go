package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Patient struct {
	// A structure to hold patient data
	gorm.Model
	ID                uint   `gorm:"primaryKey;autoIncrement"`
	uuid              string `gorm:"type:uuid;default:UUIDv7();unique;not null"`
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

func NewPatient(w http.ResponseWriter, r *http.Request) {
	var patient Patient
	err := json.NewDecoder(r.Body).Decode(&patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	patient.ID = 0
	tempUUID, _ := uuid.NewV7()
	patient.uuid = tempUUID.String()

	patient, err = CreatePatient(ctx, db, patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patient)
}

func ListPatients(w http.ResponseWriter, r *http.Request) {
	var patients []Patient
	db.Find(&patients)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patients)
}

func GetPatient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var patients []Patient
	var patient Patient
	db.Where("uuid = ?", uuid).Find(&patients)
	if len(patients) == 0 {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	} else if len(patients) == 1 {
		patient = patients[0]
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patient)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patients)
	}
}

func UpdatePatient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var patient Patient
	var patients []Patient
	err := json.NewDecoder(r.Body).Decode(&patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	db.Where("uuid = ?", uuid).Find(&patients)
	if len(patients) == 0 {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	} else if len(patients) == 1 {
		db.Model(&patients[0]).Updates(&patient)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patients[0])
	} else {
		http.Error(w, "Multiple patients found", http.StatusInternalServerError)
	}
}

func DeletePatient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var patients []Patient
	db.Where("uuid = ?", uuid).Find(&patients)
	if len(patients) == 0 {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	} else if len(patients) == 1 {
		db.Delete(&patients[0])
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patients[0])
	} else {
		http.Error(w, "Multiple patients found", http.StatusInternalServerError)
	}
}

/* TODO:
We need to implement the get appointments REST API endpoint:
* `GET /patients/:uuid/appointments` to get a list of all appointments for a particular patient.
*/
