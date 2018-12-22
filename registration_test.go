package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistration(t *testing.T) {
	t.Log("with initialized server.")
	{
		s := NewServer("127.0.0.1:8080", ioutil.Discard, testStorage())
		go func() {
			assert.Nil(t, s.ListenAndServe())
		}()
		defer s.Close()

		t.Log("\ttest:0\tshould return bad request if the body is invalid.")
		{
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/registrate", s.Addr), strings.NewReader("invalid"))
			assert.Nil(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}

		t.Log("\ttest:1\tshould registrate user with valid body.")
		{
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/registrate", s.Addr), strings.NewReader(`{"email":"new@domain.zone", "password": "qwerty", "password_confirmation": "qwerty"}`))
			assert.Nil(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		t.Log("\ttest:2\tshould validate email uniqueness.")
		{
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/registrate", s.Addr), strings.NewReader(`{"email":"exists@domain.zone"}`))
			assert.Nil(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		}

		t.Log("\ttest:3\tshould validate password and password_confirmation match.")
		{
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/registrate", s.Addr), strings.NewReader(`{"email":"new@domain.zone", "password": "qwerty", "password_confirmation": "other"}`))
			assert.Nil(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		}

		t.Log("\ttest:4\tshould validate email.")
		{
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/registrate", s.Addr), strings.NewReader(`{"email":"invalid", "password": "qwerty", "password_confirmation": "qwerty"}`))
			assert.Nil(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		}
	}
}

func testStorage() *MemStore {
	repo := MemStore{
		Users: []User{
			User{
				ID:       1,
				Email:    "exists@domain.zone",
				Password: "qwerty",
			},
		},
	}

	return &repo
}
