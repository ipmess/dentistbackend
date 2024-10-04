package patient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ipmess/dentistbackend/pkg/models"
	"gorm.io/gorm"
)

type HTTPHandler struct {
	DB  *gorm.DB
	Ctx context.Context
}

// CreatePatient creates a new patient record in the database
func CreatePatient(ctx context.Context, db *gorm.DB, patient models.Patient) (models.Patient, error) {
	if err := db.Create(&patient).Error; err != nil {
		return models.Patient{}, err
	}
	return patient, nil
}

// PrintPatient prints the patient details
func PrintPatient(patient models.Patient) {
	patient_json, _ := json.MarshalIndent(patient, "", "  ")
	fmt.Println(string(patient_json))
}

func (h *HTTPHandler) NewPatient(w http.ResponseWriter, r *http.Request) {
	var patient models.Patient
	err := json.NewDecoder(r.Body).Decode(&patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	patient.ID = 0
	tempUUID, _ := uuid.NewV7()
	patient.UUID = tempUUID.String()

	patient, err = CreatePatient(h.Ctx, h.DB, patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patient)
}

func (h *HTTPHandler) ListPatients(w http.ResponseWriter, r *http.Request) {
	var patients []models.Patient
	h.DB.Find(&patients)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patients)
}

func (h *HTTPHandler) GetPatient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var patients []models.Patient
	var patient models.Patient
	h.DB.Where("uuid = ?", uuid).Find(&patients)
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

func (h *HTTPHandler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var patient models.Patient
	var patients []models.Patient
	err := json.NewDecoder(r.Body).Decode(&patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.DB.Where("uuid = ?", uuid).Find(&patients)
	if len(patients) == 0 {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	} else if len(patients) == 1 {
		h.DB.Model(&patients[0]).Updates(&patient)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patients[0])
	} else {
		http.Error(w, "Multiple patients found", http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) DeletePatient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	var patients []models.Patient
	h.DB.Where("uuid = ?", uuid).Find(&patients)
	if len(patients) == 0 {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	} else if len(patients) == 1 {
		h.DB.Delete(&patients[0])
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
