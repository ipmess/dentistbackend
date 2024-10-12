# Appointment Management Application for a Dentist

## Objective

The aim is to develop a comprehensive and user-friendly appointment management system tailored to a dentist's needs. The dentist will manage all appointments through a web application on the desktop and a native app on her iPhone.

The system will include features such as:

* synchronization with Google Calendar,
* advanced filtering,
* and data security.

The solution will not allow any Internet users to book appointments themselves. The only person that will be able to book appointments will be the dentist, after verbal communication with each patient. During this verbal communication, the dentist wants to be able to view a calendar view of her schedule so she can offer options to the patients.

## Architecture and technologies to be used in implementing this solution

### Requirements Overview:

The system will provide the dentist with the ability to:

* Manage Appointments:
   * Schedule and manage appointments, including the following details:
      * Patient name
      * Patient phone number
      * Patient email
      * Appointment type (e.g., cleaning, extraction, root canal)
      * Start time and duration

During the scheduling phase, the dentist needs to be able to quickly filter her calendar to see available time slots. The filtering needs will be, for example, "available time slots for a 40 minute appointment, during weekdays, after 16:00". Or, "1 hour appointment on Thursdays".

* Calendar Views and Customization:
   * Provide monthly, weekly, and daily calendar views.
   * Offer customizable filters to view appointments by type, patient, or date range.
   * Allow manual entry of off days (holidays or personal time) or specific off hours for rest or study.

* Patient Information  View:
   * View and edit patient details:
     * Name
     * Phone number
     * email address
     * Default notification preferences
        * via Viber
        * via email
        * via SMS
        * via WhatsApp
    * View the patient's appointment history (only date, duration, and type of appointment)

* Data Management:
   * Auto-complete patient information from the database (names, phone numbers, and email) during the appointment phase.
   * Dropdown menus for selecting appointment type.
   * The duration of each appointment is a custom value.

Each appointment type will have a default value, (for example: extraction: 30 minutes), but depending on each patient and the particular medical case, the dentist must have the capability during booking to edit this value to whatever duration she might need.

* Conflict Detection:
   * The appointment scheduling logic should not allow the dentist to book overlapping appointments.
   * Include an appointment conflict checker that alerts the dentist of scheduling conflicts (e.g., overlapping appointments) before confirming new appointments.

* Google Calendar Integration:
  * Synchronize all appointments with Google Calendar in near-real-time.
  * Detect external changes made directly in Google Calendar via Google Calendar Push Notifications (webhooks) and update the application accordingly.
  * The dentist does not currently expect to be booking appointments through Google Calendar, so this should not be supported.
  * Support at the movement of appointments around (for rescheduling, for example).

* Offline Mode for Mobile App:
  * Provide an offline mode for the mobile app, allowing the dentist to view and manage appointments without internet access. Changes made offline will synchronize with the backend when an Internet connection is restored.

* Data Security:
  * Encrypt data at rest using Transparent Data Encryption (TDE) with AWS RDS.
  * Secure data in transit via HTTPS
  * Store encryption keys using AWS Key Management Service (KMS).

## Proposed Architecture

1. Backend: REST API Built with Go

The backend will be built using Go. It will leverage REST API architecture for handling appointment and patient management tasks.
The backend will also:

* Synchronize all appointment and patient data to the database
* Synchronize the appointments with Google Calendar
* Synchronize all appointment and patient data between the web application and the mobile app
* Provide appointment and patient data to the web application and the mobile app.

## Backend Components

- **API Server**: Developed using Go frameworks like **Gin**, the API will manage the following endpoints:

#### API endpoints

##### Appointments

* `POST /appointments` to create an appointment
* `GET /appointments/month` to get a list of all appointments for a particular month/year.
* `GET /appointments/week` to get a list of all appointments for a particular week/year.
* `GET /appointments/date` to get a list of all appointments for a particular date.
* `GET /appointments/:uuid` to get a specific appointment
* `PUT /appointments/:uuid` to update a specific appointment
* `DELETE /appointments/:uuid` to delete a specific appointment

##### Patients

* `POST /patients` to create a patient
* `GET /patients` to get all patients
* `GET /patients/:uuid` to get a specific patient
* `PUT /patients/:uuid` to update a specific patient
* `GET /patients/:uuid/appointments` to get a list of all appointments for a particular patient.
* `DELETE /patients/:uuid` to delete a patient

#### API endpoint testing

The endpoint testing should test the REST API endpoints.

### Backend Features:

* Conflict Detection: Before confirming an appointment, the system will check for any time conflicts with other scheduled appointments.
* Google Calendar Integration: Integrate with the Google Calendar API to sync appointments, using push notifications (webhooks) to detect and respond to changes made externally.
* Token-Based Authentication: Secure API access using JWT tokens for authentication.
* Encryption: Store all data with AWS RDS using Transparent Data Encryption (TDE).

### Google Calendar Sync

* Synchronize all appointments with Google Calendar
* Authenticate to Google Calendar using OAuth 2.0.
* Store the OAuth 2.0 credentials on AWS KMS.
* Implement Push Notifications (webhooks) from Google Calendar to detect changes made directly on Google Calendar and sync them with the backend and in turn with the frontend components.

### Deployment of backend

* AWS EC2 or Lambda: The Go API will run on AWS.
* AWS RDS with MariaDB: Use AWS RDS for storage of appointment and patient data.

## Frontend Components

### Web Application (Desktop):

The desktop interface will be built using Next.js to provide an experience similar to Google Calendar with advanced filtering options.

* Calendar Views: Integrate the `FullCalendar` library to support monthly, weekly, and daily views.
  * Drag-and-drop rescheduling for appointments.
  * Click-to-create interactions and tooltips to display appointment details.
  * Custom Filters: The Dentist will be able to filter appointments by type, patient, or time range.
  * In calendar view, each type of appointment should be color-coded.
* Appointment Management: Forms will allow the dentist to:
  * Select a patient from a searchable dropdown.
  * Auto-populate patient details, including phone numbers and email.
  * Select appointment types and time slots with ease.

#### User Interface (UI)

The front-end will be built using Next.js, which offers server-side rendering (SSR) and static site generation (SSG) out of the box, improving both performance. The UI will feature:

* Interactive Calendar Views (monthly, weekly, daily): The dentist can view, create, and manage appointments. The UI will use components from `FullCalendar.io`, allowing drag-and-drop features to reschedule appointments easily.
* Appointment Forms: Modal-based forms for creating new appointments, complete with validation and auto-complete features for patient details.
* Responsive Design: The UI will be fully responsive and mobile-friendly, ensuring that the dentist can manage appointments from any device.

#### Routing

Next.js uses file-based routing. Each page in the pages/ directory corresponds to a route. For example:
        
* `pages/index.js`: Home or dashboard.
* `pages/appointments.js`: Calendar view of all appointments.
* `pages/patients/[id].js`: Patient details view, using dynamic routing for individual patient data.

#### Data Fetching

Next. js data fetching strategies:

* `getStaticProps`: For static patient data (e.g., general patient information), to be rendered at build time.
* `getServerSideProps`: For dynamic appointment data that changes frequently, such as the calendar view, ensuring the latest appointments are shown on every page load.

#### Rendering

* _Server-Side Rendering_ (SSR) will be used for pages like the calendar and patient dashboard, ensuring fast load times and SEO benefits.
* _Client-Side Rendering_ (CSR) will handle real-time updates (like drag-and-drop appointments) without reloading the entire page.
* _Static Site Generation_ (SSG) will be used for pages that do not frequently change, such as patient profiles or general information.

#### Integrations

The frontend Next.js application will integrate with the following components:

* The Backend
* Authentication?

#### Performance

To optimize the Frontend performance, several techniques will be employed:

* Code splitting: Next.js automatically splits the code, loading only the necessary JS and CSS for the current page, reducing initial load times.
* Lazy loading: Components such as patient search results and large appointment calendars will be lazily loaded, improving performance.

#### Developer Experience

* TypeScript will be used for the front-end code, increasing type safety.
* Linting and Prettier will be integrated to maintain clean and consistent code.
* ESLint will be used to enforce coding standards on the frontend application.
* Docker: Eventually, a Dockerized development environment should be developed.  A docker environment will provide a more consistent behavior between the development and production environments.

### Mobile Application:

The mobile app will be developed using React Native with Expo to streamline development and ensure cross-platform compatibility (iOS and Android).

* Calendar Views: Similar to the desktop, a mobile-friendly version using libraries like react-native-calendars.
* Offline Mode: The mobile app will include an offline mode, allowing the dentist to access and manage appointments without an internet connection. Changes will sync when reconnected to the backend.
* Optimized Touch Experience: Forms and interaction patterns will be designed for ease of use on touch devices.

#### State Management:

* Redux will be used for state management, to maintain a consistent state across both web and mobile applications.

## Database and Synchronization

### Database:

The backend will communicate with MariaDB running on AWS RDS. The MariaDB database will serve as the central database.

#### Database Schema:
      
* Patients Table: Stores patient information
   * patient_id
   * patient_UUID
   * name
   * phone_number
   * email
   * Default_contact_preferences
      * Viber
      * WhatsApp
      * SMS
      * email
      * hours before appointment time
      * Every how many days since the last appointment does the patient want to be reminded that they should come back for a visit
   * default notification preferences since last appointment (days)

* Appointments Table: Stores appointment details
   * ID (primary key, unique ID per appointment)
   * appointment_UUID
   * patient_id (which references the patient_id of the Patients table)
   * Appointment type ID (which references the Appointment type_id of the AppointmentType table)
   * start_date_time
   * duration
   * Viber (whether the patient wants to notified about this appointment via Viber)
   * WhatsApp (whether the patient wants to notified about this appointment via WhatsApp)
   * email (whether the patient wants to notified about this appointment via email)
   * SMS (whether the patient wants to notified about this appointment via SMS)
   * Reminder (the number of hours before the appointment that the patient wants to be remind

* Appointment Type table: stores a description and some defaults per appointment type
   * ID (primary key, unique ID per appointment type)
   * Description (string describing the type of appointment (extraction, filling, cleaning, whitening, etc))
   * Default Duration (in minutes)
   * Color (e.g. #FFA07A, used as a color in calendar view for all appointments of this type)

## User Experience (UX)

To ensure the system is intuitive and easy to use, the following UX features will be included:

* Google Calendar-Like Interface: Both web and mobile interfaces will resemble Google Calendar, providing familiar navigation and scheduling features.
* Custom Filters: The ability to filter appointments by patient, type, or time range enhances usability.
* Appointment Conflict Alerts: An automatic checker for overlapping appointments ensures that the dentist won’t inadvertently schedule conflicting appointments.
* Manual Off Days/Hours Entry: The dentist can easily schedule time off or block off specific hours for personal reasons, which will be reflected across all platforms and in Google Calendar.

## Deployment Plan

### Frontend Deployment:

* Next.js Web App: Deploy the web-based frontend via Netlify or on an AWS EC2 instance. If it is necessary, host the Next.js app behind Nginx.
* React Native Mobile App: Deploy the mobile app using Expo, and submit it to the Apple App Store.

### Backend Deployment:

* Go REST API: Hosted on AWS Lambda (serverless) or AWS EC2.
* Database: MariaDB on AWS RDS. Deploy Transparent Data Encryption (TDE) on RDS.

## Work-in-progress notes

My development environment includes a locally running MariaDB, with a dentist user and a database for this project. Steps to install a local instance of MariaDB, create a user, and a database:

```
sudo apt-install mariadb-server
```

After you install the mariadb server on the local machine, confirm that it is running:

```
sudo systemctl status mariadb
```

(it should say `active (running)`)

Next, let's connect to the MariDB instance:

```
$ sudo mariadb -uroot -p -h localhost
```

You should get a MariaDB prompt:

```
MariaDB [(none)]>
```

In the MariaDB CLI, enter the following commands:

**Create a user for the application:**

```
CREATE USER `dentist`@`%` IDENTIFIED BY `somelamepass`;
```
**Create a database for the application:**

```
CREATE DATABASE appointments_db;
```
**Grant rights to the new database to the application's user:**

```
GRANT ALL PRIVILEGES ON appointments_db.* TO `dentist`@`%` WITH GRANT OPTION;
```

Use the `quit` command in MariaDB to leave the MariaDB CLI. From your host, confirm you can connect to the newly created database using the new user's credentials:

```
$mariadb -u dentist -p appointments_db
```

And you should now get a (different) prompt from MariaDB:

```
MariaDB [appointments_db]>
```

#### JSON config file

Once you have the development database in place, you can now configure the `config.json` file that the application is expecting so you can connect to this database.

**Sample `config.json`:**

```
{
   "db_endpoint": "localhost:3306",
   "username": "dentist",
   "password": "somelamepass",
   "database": "appointments_db"
   "populate_db": true
}
```

#### To clear the current database and re-populate it with sample data:

To connect to the local database on the local host, run:

```
mariadb -udentist -p appointments_db
```

The password for the _dentist_ user is in the `config.json` file.

Once in MariaDB, run the following to verify that all the tables are there:

```
MariaDB [appointments_db]> show tables;
+---------------------------+
| Tables_in_appointments_db |
+---------------------------+
| appointment_types         |
| appointments              |
| patients                  |
+---------------------------+
3 rows in set (0.001 sec)
```

To drop all tables, run:

```
MariaDB [appointments_db]> drop table appointments;
Query OK, 0 rows affected (0.448 sec)

MariaDB [appointments_db]> drop table appointment_types;
Query OK, 0 rows affected (0.169 sec)

MariaDB [appointments_db]> drop table patients;
Query OK, 0 rows affected (0.227 sec)

MariaDB [appointments_db]> show tables;
Empty set (0.001 sec)

```