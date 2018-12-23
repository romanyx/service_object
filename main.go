package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

const (
	passwordMismatch = "password mismatch"
	emailExists      = "email exists"
	validationMsg    = "you have validation errors"
)

var (
	// ErrEmailExists returns when given email is present
	// in storage.
	ErrEmailExists = errors.New("email already exists")
)

func main() {
	var (
		addr  = flag.String("addr", ":8080", "address of the http server")
		debug = flag.Bool("debug", false, "enable debug")
	)

	stdout := ioutil.Discard
	if *debug {
		stdout = os.Stdout
	}

	r := MemStore{}
	s := NewServer(*addr, stdout, &r)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

// NewServer prepares http server.
func NewServer(addr string, stdout io.Writer, r Repository) *http.Server {
	mux := http.NewServeMux()
	srv := &Service{
		Validater: &PlayValidator{
			Validater:  validator.New(),
			Repository: r,
		},
		Repository: r,
	}

	h := RegistrationHandler{
		Registrater: NewRegistraterWithLog(srv, stdout, os.Stderr),
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
	Unique(ctx context.Context, email string) error
	Create(context.Context, *Form) (*User, error)
}

// Validater validation abstraction.
type Validater interface {
	Validate(context.Context, *Form) error
}

// ValidationErrors holds validation errors.
type ValidationErrors map[string]string

// Error implements error interface.
func (v ValidationErrors) Error() string {
	return validationMsg
}

// Service holds data required for registration.
type Service struct {
	Validater
	Repository
}

// Registrate holds registration domain logic.
func (s *Service) Registrate(ctx context.Context, f *Form) (*User, error) {
	if err := s.Validater.Validate(ctx, f); err != nil {
		return nil, errors.Wrap(err, "validater validate")
	}

	user, err := s.Create(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "repository create")
	}

	return user, nil
}

// Registrater abstraction for registration service.
type Registrater interface {
	Registrate(context.Context, *Form) (*User, error)
}

// RegistrationHandler for regisration requests.
type RegistrationHandler struct {
	Registrater
}

// ServerHTTP implements http.Handler.
func (h *RegistrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var f Form
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.Registrate(r.Context(), &f)
	if err != nil {
		switch v := errors.Cause(err).(type) {
		case ValidationErrors:
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(v)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
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

// Unique checks if a email exists in the database.
func (s *MemStore) Unique(ctx context.Context, email string) error {
	for _, u := range s.Users {
		if u.Email == email {
			return ErrEmailExists
		}
	}

	return nil
}

// Create creates user in the database for a form.
func (s *MemStore) Create(ctx context.Context, f *Form) (*User, error) {
	u := User{
		ID:       len(s.Users) + 1,
		Password: f.Password,
		Email:    f.Email,
	}

	s.Users = append(s.Users, u)

	return &u, nil
}

// PlayValidator holds registration form validations.
type PlayValidator struct {
	Validater *validator.Validate
	Repository
}

// Validate implements Validater.
func (v *PlayValidator) Validate(ctx context.Context, f *Form) error {
	validations := make(ValidationErrors)

	err := v.Validater.Struct(f)
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

	if err := v.Repository.Unique(ctx, f.Email); err != nil {
		if err != ErrEmailExists {
			return errors.Wrap(err, "repository unique")
		}

		validations["email"] = emailExists
	}

	if len(validations) > 0 {
		return validations
	}

	return nil
}
