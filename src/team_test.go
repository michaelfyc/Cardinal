package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// three team accounts
// vidar (change name to Vidar)
// e99 (login)
// John	(delete)

func TestService_NewTeams(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name": "vidar",
		"Logo": "",
	})
	req, _ := http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Logo": "",
	}})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// repeat in form
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Name": "vidar",
		"Logo": "",
	}, {
		"Name": "vidar",
		"Logo": "test",
	}})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Name": "vidar",
		"Logo": "",
	}, {
		"Name": "E99",
		"Logo": "test_image.png",
	}, {
		"Name": "John",
		"Logo": "test_image123.png",
	},
	})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// save the team password
	var password struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  []struct {
			Name     string `json:"Name"`
			Password string `json:"Password"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &password)
	assert.Equal(t, nil, err)
	teamPassword = append(teamPassword, password.Data...)

	// repeat in database
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Name": "vidar",
		"Logo": "",
	}, {
		"Name": "E99",
		"Logo": "test_image.png",
	}})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestService_GetAllTeams(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/manager/teams", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_GetTeamInfo(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/manager/teams", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_EditTeam(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name": "vidar",
		"Logo": "",
	})
	req, _ := http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   233,
		"Name": "vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// team name repeat
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   2,
		"Name": "vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   1,
		"Name": "Vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_ResetTeamPassword(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"IDd": 3,
	})
	req, _ := http.NewRequest("POST", "/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID": 233,
	})
	req, _ = http.NewRequest("POST", "/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID": 3,
	})
	req, _ = http.NewRequest("POST", "/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_DeleteTeam(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/manager/team?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/manager/team?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/manager/team?id=3", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_TeamLogin(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name":     123123,
		"Password": "",
	})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error password
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     teamPassword[2].Name,
		"Password": "aaa",
	})
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     teamPassword[1].Name,
		"Password": teamPassword[1].Password,
	})
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_TeamLogout(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logout", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}