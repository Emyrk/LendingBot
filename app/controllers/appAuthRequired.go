package controllers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Emyrk/LendingBot/app/core/cryption"
	"github.com/revel/revel"
)

type AppAuthRequired struct {
	*revel.Controller
}

type Enable2fa struct {
	Pass   string `json:"pass"`
	Enable bool   `json:"enable"`
	Token  string `json:"token"`
}

type Pass struct {
	Pass string `json:"pass"`
}

func (r AppAuthRequired) unmarshal2fa(body io.ReadCloser) *Enable2fa {
	var json2fa Enable2fa
	err := json.NewDecoder(body).Decode(&json2fa)
	if err != nil {
		fmt.Printf("Error unmarshaling pass: %s", err.Error())
		return &json2fa
	}
	defer body.Close()
	return &json2fa
}

func (r AppAuthRequired) unmarshalPass(body io.ReadCloser) string {
	var jsonPass Pass
	err := json.NewDecoder(body).Decode(&jsonPass)
	if err != nil {
		fmt.Printf("Error unmarshaling pass: %s", err.Error())
		return ""
	}
	defer body.Close()
	return jsonPass.Pass
}

func (r AppAuthRequired) Dashboard() revel.Result {
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	email, _ := cryption.VerifyJWT(tokenString, state.JWTSecret)
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Println("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}
	r.ViewArgs["UserLevel"] = fmt.Sprintf("%d", u.Level)
	return r.Render()
}

func (r AppAuthRequired) Logout() revel.Result {
	r.Session[cryption.COOKIE_JWT_MAP] = ""
	return r.Redirect(App.Index)
}

func (r AppAuthRequired) InfoDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/InfoDashboard.html")
}

func (r AppAuthRequired) Enable2FA() revel.Result {
	data := make(map[string]interface{})
	json2fa := r.unmarshal2fa(r.Request.Body)
	if json2fa == nil {
		fmt.Printf("Error grabbing 2fa err\n")
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	email, _ := cryption.VerifyJWT(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)

	err := state.Enable2FA(email, json2fa.Pass, json2fa.Token, json2fa.Enable)
	if err != nil {
		fmt.Printf("Error enabling 2fa err: %s\n", err.Error())
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	u, _ := state.UserDB.FetchUser(email)
	data[JSON_DATA] = fmt.Sprintf("%t", u.Enabled2FA)

	return r.RenderJSON(data)
}

func (r AppAuthRequired) InfoAdvancedDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/InfoAdvancedDashboard.html")
}

func (r AppAuthRequired) SettingsDashboard() revel.Result {
	email, _ := cryption.VerifyJWT(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)
	u, _ := state.FetchUser(email)

	r.ViewArgs["has2FA"] = fmt.Sprintf("%t", u.Has2FA)
	r.ViewArgs["enabled2FA"] = fmt.Sprintf("%t", u.Enabled2FA)

	return r.RenderTemplate("AppAuthRequired/SettingsDashboard.html")
}

func (r AppAuthRequired) Create2FA() revel.Result {
	pass := r.unmarshalPass(r.Request.Body)
	email, _ := cryption.VerifyJWT(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)

	data := make(map[string]interface{})

	qr, err := state.Add2FA(email, pass)
	if err != nil {
		fmt.Printf("Error authenticating 2fa err: %s\n", err.Error())
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	data[JSON_DATA] = qr

	return r.RenderJSON(data)
}

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	email, err := cryption.VerifyJWT(tokenString, state.JWTSecret)
	if err != nil {
		fmt.Printf("WARNING: AuthUser failed JWT Token: [%s] and error: %s\n", tokenString, err.Error())
		return r.Redirect(App.Index)
	}

	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Printf("WARNING: AuthUser failed to fetch user: %s\n", tokenString)
		return r.Redirect(App.Index)
	}

	return nil
}
