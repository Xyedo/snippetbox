package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/xyedo/snippetbox/internal/assert"
)

func TestViewHome(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	code, _, body := ts.get(t, "/")
	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, string(body), "<h2>Latest Snippets</h2>")
}
func TestSnippetView(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			path:     "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name:     "Non-existent ID",
			path:     "/snippet/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			path:     "/snippet/view/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			path:     "/snippet/view/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			path:     "/snippet/view/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			path:     "/snippet/view",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.path)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, string(body), tt.wantBody)
			}
		})
	}
}
func TestUserSignup(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)
	const (
		validName     = "Bob"
		validPassword = "validPa$$word"
		validEmail    = "bob@example.com"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)
	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		}, {
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, "/user/signup", form)
			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, string(body), tt.wantFormTag)
			}
		})
	}
}
func TestAccountView(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	code, headers, _ := ts.get(t, "/account/view")
	assert.Equal(t, code, http.StatusSeeOther)
	assert.Equal(t, headers.Get("Location"), "/user/login")

	code, _, body := ts.get(t, "/user/login")
	assert.Equal(t, code, http.StatusOK)
	csrfToken := extractCSRFToken(t, body)
	form := url.Values{}
	form.Add("email", "alice@example.com")
	form.Add("password", "pa$$word")
	form.Add("csrf_token", csrfToken)
	code, headers, _ = ts.postForm(t, "/user/login", form)
	assert.Equal(t, code, http.StatusSeeOther)
	assert.Equal(t, headers.Get("Location"), "/account/view")
	code, _, body = ts.get(t, "/account/view")
	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, string(body), `<h2>Your Account</h2>`)
}

func TestCreateSnippet(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	t.Run("Unauthenticated", func(t *testing.T) {
		code, headers, _ := ts.get(t, "/snippet/create")
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/user/login")
	})
	t.Run("Authenticated", func(t *testing.T) {
		code, _, body := ts.get(t, "/user/login")
		assert.Equal(t, code, http.StatusOK)
		csrfToken := extractCSRFToken(t, body)
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "pa$$word")
		form.Add("csrf_token", csrfToken)
		code, headers, _ := ts.postForm(t, "/user/login", form)
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/snippet/create")
		code, _, body = ts.get(t, "/snippet/create")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, string(body), `<form action='/snippet/create' method='POST'>`)
		const (
			validTitle   = "MANTAP"
			validContent = "asik asik jos! -Hafid Mahdi"
			validExpires = "365"
		)
		tests := []struct {
			name      string
			title     string
			content   string
			expires   string
			csrfToken string
			wantCode  int
			wantBody  string
		}{
			{
				name:      "Valid Create Snippet",
				title:     validTitle,
				content:   validContent,
				expires:   validExpires,
				csrfToken: csrfToken,
				wantCode:  http.StatusSeeOther,
			},
			{
				name:      "Blank Title",
				title:     "",
				content:   validContent,
				expires:   validExpires,
				csrfToken: csrfToken,
				wantCode:  http.StatusUnprocessableEntity,
				wantBody:  "This field cannot be blank",
			},
			{
				name:      "Title is More than 100 chars",
				title:     "Skill untuk jadiin obrolan biasa menjadi meaningful conversation itu emg sulit banget didapetindfsdfsfdddd",
				content:   validContent,
				expires:   validExpires,
				csrfToken: csrfToken,
				wantCode:  http.StatusUnprocessableEntity,
				wantBody:  "This filed cannot be more than 100 characters long",
			},
			{
				name:      "Blank Content",
				title:     validContent,
				content:   "",
				expires:   validExpires,
				csrfToken: csrfToken,
				wantCode:  http.StatusUnprocessableEntity,
				wantBody:  "This field cannot be blank",
			},
			{
				name:      "Invalid Value",
				title:     validContent,
				content:   validContent,
				expires:   "255",
				csrfToken: csrfToken,
				wantCode:  http.StatusUnprocessableEntity,
				wantBody:  "This field must equal 1, 7, or 365",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				form := url.Values{}
				form.Add("title", tt.title)
				form.Add("content", tt.content)
				form.Add("expires", tt.expires)
				form.Add("csrf_token", tt.csrfToken)
				code, _, body := ts.postForm(t, "/snippet/create", form)
				assert.Equal(t, code, tt.wantCode)

				if tt.wantBody != "" {
					assert.StringContains(t, string(body), tt.wantBody)
				}
			})
		}
	})
}
func TestUserLogin(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	code, _, body := ts.get(t, "/user/login")
	validCSRFToken := extractCSRFToken(t, body)
	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, string(body), `<form action='/user/login' method='POST' novalidate>`)
	const (
		validEmail    = "alice@example.com"
		validPassword = "pa$$word"
	)
	tests := []struct {
		name      string
		email     string
		password  string
		csrfToken string
		wantCode  int
		wantBody  string
	}{
		{
			name:      "Valid submission",
			email:     validEmail,
			password:  validPassword,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusSeeOther,
		},
		{
			name:      "Blank Email",
			email:     "",
			password:  validPassword,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
			wantBody:  "This field cannot be blank",
		},
		{
			name:      "Invalid email",
			email:     "hasda",
			password:  validPassword,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
			wantBody:  "This field must be a valid email address",
		},
		{
			name:      "Blank Password",
			email:     validEmail,
			password:  "",
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
			wantBody:  "This field cannot be blank",
		},
		{
			name:      "Invalid Credentials",
			email:     "hafid@gmail.com",
			password:  validPassword,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
			wantBody:  "Email or password is incorrect",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("email", tt.email)
			form.Add("password", tt.password)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, "/user/login", form)
			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, string(body), tt.wantBody)
			}
		})
	}
}

func TestUserLogout(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	const (
		validEmail    = "alice@example.com"
		validPassword = "pa$$word"
	)

	t.Run("Unauthenticated", func(t *testing.T) {
		code, _, body := ts.get(t, "/user/login")
		assert.Equal(t, code, http.StatusOK)
		validCSRFToken := extractCSRFToken(t, body)
		form := url.Values{}
		form.Add("csrf_token", validCSRFToken)
		code, headers, _ := ts.postForm(t, "/user/logout", form)
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/user/login")
	})
	t.Run("Authenthticated", func(t *testing.T) {
		code, _, body := ts.get(t, "/user/login")
		validCSRFToken := extractCSRFToken(t, body)
		assert.Equal(t, code, http.StatusOK)
		form := url.Values{}
		form.Add("email", validEmail)
		form.Add("password", validPassword)
		form.Add("csrf_token", validCSRFToken)
		code, _, _ = ts.postForm(t, "/user/login", form)
		assert.Equal(t, code, http.StatusSeeOther)
		form = url.Values{}
		form.Add("csrf_token", validCSRFToken)
		code, _, _ = ts.postForm(t, "/user/logout", form)
		assert.Equal(t, code, http.StatusSeeOther)
	})
}
func TestUserChangePassword(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	const (
		validEmail              = "alice@example.com"
		validCurrPassword       = "pa$$word"
		validNewPassword        = "qwerty123"
		validNewConfirmPassword = validNewPassword
	)
	code, _, body := ts.get(t, "/user/login")
	assert.Equal(t, code, http.StatusOK)
	validCSRFToken := extractCSRFToken(t, body)
	form := url.Values{}
	form.Add("email", validEmail)
	form.Add("password", validCurrPassword)
	form.Add("csrf_token", validCSRFToken)
	code, _, _ = ts.postForm(t, "/user/login", form)
	assert.Equal(t, code, http.StatusSeeOther)
	code, _, body = ts.get(t, "/account/password/update")
	assert.Equal(t, code, http.StatusOK)
	validCSRFToken = extractCSRFToken(t, body)
	tests := []struct {
		name               string
		currentPassword    string
		newPassword        string
		newpasswordConfirm string
		validcsrf          string
		wantCode           int
		wantBody           string
	}{
		{
			name:               "Valid Change Password",
			currentPassword:    validCurrPassword,
			newPassword:        validNewPassword,
			newpasswordConfirm: validNewConfirmPassword,
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusSeeOther,
		},
		{
			name:               "Blank Current Password",
			currentPassword:    "",
			newPassword:        validNewPassword,
			newpasswordConfirm: validNewConfirmPassword,
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusUnprocessableEntity,
			wantBody:           "This field cannot be blank",
		},
		{
			name:               "Blank New Password",
			currentPassword:    validCurrPassword,
			newPassword:        "",
			newpasswordConfirm: validNewConfirmPassword,
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusUnprocessableEntity,
			wantBody:           "This field cannot be blank",
		},
		{
			name:               "Blank New Password Confirm",
			currentPassword:    validCurrPassword,
			newPassword:        validNewPassword,
			newpasswordConfirm: "",
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusUnprocessableEntity,
			wantBody:           "This field cannot be blank",
		},
		{
			name:               "New Password less than 8 chars",
			currentPassword:    validCurrPassword,
			newPassword:        "123",
			newpasswordConfirm: "123",
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusUnprocessableEntity,
			wantBody:           "This field must be at least 8 characters long",
		},
		{
			name:               "New Password Not match with ConfirmPass",
			currentPassword:    validCurrPassword,
			newPassword:        validCurrPassword,
			newpasswordConfirm: "123",
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusUnprocessableEntity,
			wantBody:           "New Password do not match",
		},
		{
			name:               "Current Password do not match",
			currentPassword:    "qwertyuiop",
			newPassword:        validNewPassword,
			newpasswordConfirm: validNewConfirmPassword,
			validcsrf:          validCSRFToken,
			wantCode:           http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("currentPassword", tt.currentPassword)
			form.Add("newPassword", tt.newPassword)
			form.Add("newPasswordConfirmation", tt.newpasswordConfirm)
			form.Add("csrf_token", tt.validcsrf)
			code, _, body := ts.postForm(t, "/account/password/update", form)
			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, string(body), tt.wantBody)
			}
		})
	}
}

func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	code, _, body := ts.get(t, "/ping")
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}
	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}
