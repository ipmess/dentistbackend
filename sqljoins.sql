SELECT appointments.ID, Appointment.PatientID,Appointment.StartTime, Appointment.Duration, Appointment.Viber, Appointment.Whatsapp,Appointment.SMS, Appointment.EmailNotification, Appointment.Reminder, Patient.Name AS PatientName,Patient.PhoneNumber, AppointmentType.Description, AppointmentType. Color FROM `appointments` left join Patient onPatient.ID = Appointment.PatientID left join AppointmentType.ID = Appointment.AppointmentTypeID


SELECT appointments.ID, appointments.Patient_ID, appointments.Start_Time, 
       appointments.Duration, appointments.viber, appointments.whatsapp, 
       appointments.SMS, appointments.Email_Notification, appointments.reminder, 
       patients.name AS PatientName, patients.phone_Number, 
       appointment_types.description, appointment_types.color 
FROM `appointments`
JOIN patients
  ON patients.ID = appointments.patient_is
JOIN appointment_types
  ON appointment_types.id = appointments.appointment_type_ID;

SELECT appointments.ID, appointments.Patient_ID, appointments.Start_Time, appointments.Duration, appointments.viber, appointments.whatsapp, appointments.SMS FROM `appointments`;

SELECT appointments.ID, appointments.Patient_ID, appointments.Start_Time, 
       appointments.Duration, appointments.viber, appointments.whatsapp, 
       appointments.SMS, appointments.Email_Notification as email, appointments.reminder, 
       patients.name AS PatientName, patients.phone_Number,
       appointment_types.description, appointment_types.color
FROM appointments
JOIN patients
  ON patients.ID = appointments.patient_id
JOIN appointment_types
  ON appointment_types.id = appointments.appointment_type_id;
  
SELECT appointments.ID, appointments.Patient_ID,appointments.Start_Time,
       appointments.Duration, appointments.viber, appointments.whatsapp,
       appointments.SMS, appointments.email_Notification, appointments.reminder,
       patients.Name AS PatientName, patients.PhoneNumber,
       appointment_types.description, appointment_types.Color
FROM `appointments` 
join patients
ON patients.ID = appointments.patient_ID
join appointment_types
ON appointment_types.ID = appointments.Appointment_Type_ID

  
