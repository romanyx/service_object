package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"gopkg.in/go-playground/validator.v9"
)

const (
	passwordMismatch = "password mismatch"
	emailExists      = "email exists"
)

func main() {
	var (
		addr = flag.String("addr", ":8080", "address of the http server")
	)

	r := MemStore{}
	s := NewServer(*addr, &r)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

// NewServer prepares http server.
func NewServer(addr string, r Repository) *http.Server {
	mux := http.NewServeMux()
	h := RegistrationHandler{
		Validator:  validator.New(),
		Repository: r,
	}

	mux.Handle("/registrate", &h)

	s := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &s
}

// Repository is a data access layer.
type Repository interface {
	Exists(email string) (bool, error)
	Create(*Form) (*User, error)
}

// RegistrationHandler for regisration requests.
type RegistrationHandler struct {
	Validator *validator.Validate
	Repository
}

// ServerHTTP implements http.Handler.
func (h *RegistrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var f Form
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	validations := make(map[string]string)

	err := h.Validator.Struct(f)
	if err != nil {
		if vs, ok := err.(validator.ValidationErrors); ok {
			for _, v := range vs {
				validations[v.Tag()] = fmt.Sprintf("%s is invalid", v.Tag())
			}
		}
	}

	if f.Password != f.PasswordConfirmation {
		validations["password"] = passwordMismatch
	}

	exists, err := h.Exists(f.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		validations["email"] = emailExists
	}

	if len(validations) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(validations)
		return
	}

	u, err := h.Create(&f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&u)
}

// Form is a regisration request.
type Form struct {
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"gte=3,lte=16"`
	PasswordConfirmation string `json:"password_confirmation" validate:"gte=3,lte=16"`
}

// User represents the database column.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// MemStore is a memroy storage for users.
type MemStore struct {
	Users []User
}

// Exists checks if a email exists in the database.
func (s *MemStore) Exists(email string) (bool, error) {
	for _, u := range s.Users {
		if u.Email == email {
			return true, nil
		}
	}

	return false, nil
}

// Create creates user in the database for a form.
func (s *MemStore) Create(f *Form) (*User, error) {
	u := User{
		ID:       len(s.Users) + 1,
		Password: f.Password,
		Email:    f.Email,
	}

	s.Users = append(s.Users, u)

	return &u, nil
}
