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
	if err != nil {
		fmt.Printf("Error authenticating err: %s\n", err.Error())
		data[JSON_ERROR] = "Error login"
		c.Response.Status = 500
		return c.RenderJSON(data)
	}
	if !ok {
		fmt.Printf("Error authenticating email: %s pass: %s\n", email, pass)
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
