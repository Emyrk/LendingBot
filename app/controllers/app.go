package controllers

import (
	"encoding/json"
	"fmt"

	"io"

	"github.com/Emyrk/LendingBot/app/core"
	"github.com/Emyrk/LendingBot/app/core/cryption"
	"github.com/Emyrk/LendingBot/app/lender"
	"github.com/Emyrk/LendingBot/app/queuer"
	"github.com/revel/revel"

	// For Prometheus
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

var state *core.State

func init() {
	// Prometheus
	lender.RegisterPrometheus()
	queuer.RegisterPrometheus()

	state = core.NewState()
	lenderBot := lender.NewLender(state)
	queuerBot := queuer.NewQueuer(state, lenderBot)

	return
	// Start go lending
	go lenderBot.Start()
	go queuerBot.Start()
	go launchPrometheus(9911)
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type JSONUser struct {
	Email string `json:"email"`
	Pass  string `json:"pass"`
}

const (
	JSON_DATA  = "data"
	JSON_ERROR = "error"
)

type App struct {
	*revel.Controller
}

type AppAuthRequired struct {
	*revel.Controller
}

func (c App) Sandbox() revel.Result {
	return c.RenderTemplate("App/Index.html")
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) unmarshalUser(body io.ReadCloser) (string, string) {
	var jsonUser JSONUser
	err := json.NewDecoder(body).Decode(&jsonUser)
	if err != nil {
		fmt.Printf("Error unmarshaling user %s", err.Error())
		return "", ""
	}
	defer body.Close()
	return jsonUser.Email, jsonUser.Pass
}

func (c App) Login() revel.Result {
	email, pass := c.unmarshalUser(c.Request.Body)

	data := make(map[string]interface{})

	ok, _, err := state.AuthenticateUser(email, pass)
	if !ok {
		fmt.Printf("Error authenticating user: %s\n", err)
		data[JSON_ERROR] = "Invalid login"
		c.Response.Status = 400
		return c.RenderJSON(data)
	}

	stringToken, err := cryption.NewJWT(email, state.JWTSecret, cryption.JWT_EXPIRY_TIME)
	if err != nil {
		data[JSON_ERROR] = fmt.Sprintf("Unable to create JWT: %s", err.Error())
		c.Response.Status = 500
		return c.RenderJSON(data)
	}

	c.Session[cryption.COOKIE_JWT_MAP] = stringToken

	return c.RenderJSON(data)
}

func (c App) Register() revel.Result {
	email, pass := c.unmarshalUser(c.Request.Body)

	data := make(map[string]interface{})

	err := state.NewUser(email, pass)
	if err != nil {
		data[JSON_ERROR] = fmt.Sprintf("Unable to create new user: %s", err.Error())
		c.Response.Status = 400
		return c.RenderJSON(data)
	}

	stringToken, err := cryption.NewJWT(email, state.JWTSecret, cryption.JWT_EXPIRY_TIME)
	if err != nil {
		data[JSON_ERROR] = fmt.Sprintf("Unable to create JWT: %s", err.Error())
		c.Response.Status = 500
		return c.RenderJSON(data)
	}

	c.Session[cryption.COOKIE_JWT_MAP] = stringToken

	return c.RenderJSON(data)
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

func (r AppAuthRequired) InfoAdvancedDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/InfoAdvancedDashboard.html")
}

func (r AppAuthRequired) SettingsDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/SettingsDashboard.html")
}

func (r AppAuthRequired) SysAdminDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/SysAdminDashboard.html")
}

func (r AppAuthRequired) AdminDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/AdminDashboard.html")
}

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	_, err := cryption.VerifyJWT(tokenString, state.JWTSecret)
	if err != nil {
		fmt.Printf("WARNING: AuthUser failed JWT Token: %s\n", tokenString)
		return r.Redirect(App.Index)
	}
	return nil
}
