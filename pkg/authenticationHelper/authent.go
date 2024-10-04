package authenticationHelper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var userCredentials = Credentials{
	Username: "dentist",
	Password: "dental_secret",
}

var jwtKey = []byte("JWT_secret")

type DentistAPIClaims struct {
	User string `json:"User"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(username string) (string, error) {
	// Create claims with multiple fields populated
	claims := DentistAPIClaims{
		username,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dentistbackend",
			Subject:   "Dentist Backend API",
			ID:        "1",
		},
	}

	// Create a new token object, specifying signing method and claims:
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedTokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("couldn't generate JWT signature: %s\n", err)
		//log.Printf("Token: %v\n", token) //not sure if this will work or cause a runtime error
		return "", err
	}
	return signedTokenString, nil
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user
	var suppliedCredentials Credentials
	err := json.NewDecoder(r.Body).Decode(&suppliedCredentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the username and password are correct:
	if userCredentials.Username == suppliedCredentials.Username {
		if userCredentials.Password == suppliedCredentials.Password {
			// Generate JWT token
			token, err := GenerateJWT(suppliedCredentials.Username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(token)
			// TODO: Examine if we should embed the token into a JSON object
			// And the JWT token should be sent in a id_token field
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func ValidateJWT(tokenString string) (DentistAPIClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &DentistAPIClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		log.Printf("couldn't parse JWT token: %s\n", err)
		return DentistAPIClaims{}, err
	}

	// Check the signing method:
	signingMethod, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok {
		return DentistAPIClaims{}, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	if signingMethod != jwt.SigningMethodHS512 {
		return DentistAPIClaims{}, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	log.Printf("Managed to parse JWT token and the signing method is HS512\n")

	// Extract the claims
	claims, ok := token.Claims.(*DentistAPIClaims)
	if !ok {
		log.Printf("couldn't extract claims from JWT token\n")
		return DentistAPIClaims{}, err
	}

	// Check whether the token is valid:
	if !token.Valid {
		return DentistAPIClaims{}, err
	}
	log.Printf("Extracted claims and the JWT token seems to be valid.\n")

	if claims.User != userCredentials.Username {
		log.Printf("Bad credentials.\n")
		return DentistAPIClaims{}, fmt.Errorf("bad credentials")
	}

	if claims.Issuer != "dentistbackend" {
		log.Printf("Bad issuer.\n")
		return DentistAPIClaims{}, fmt.Errorf("bad issuer")
	}

	if claims.Subject != "Dentist Backend API" {
		log.Printf("Bad subject.\n")
		return DentistAPIClaims{}, fmt.Errorf("bad subject")
	}

	issueDate, err := claims.GetIssuedAt()
	if err != nil {
		log.Printf("couldn't get issue date from JWT token")
		return DentistAPIClaims{}, err
	}
	if issueDate.After(time.Now()) {
		log.Printf("JWT token issued in the future.\n")
		return DentistAPIClaims{}, fmt.Errorf("bad JWT token issued in the future")
	}

	expiryDate, err := claims.GetExpirationTime()
	if err != nil {
		log.Printf("couldn't get expiry date from JWT token\n")
		return DentistAPIClaims{}, err
	}
	if expiryDate.Before(time.Now()) {
		log.Printf("JWT token expired.\n")
		return DentistAPIClaims{}, fmt.Errorf("expired JWT token expired")
	}
	log.Printf("Checked username, issuer, subject, issue date, and Expiration Time. Token seems valid.\n")

	return *claims, nil
}
