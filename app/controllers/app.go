package controllers

import (
	"encoding/json"
	"fmt"

	"io"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/cryption"
	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/lender"
	"github.com/Emyrk/LendingBot/src/queuer"
	"github.com/revel/revel"

	// For Prometheus
	"github.com/prometheus/client_golang/prometheus"
	"net/http"

	// Init logger
	_ "github.com/Emyrk/LendingBot/src/log"
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

func MakeURL(safeUrl string) string {
	if revel.DevMode {
		return "http://localhost:9000/" + safeUrl
	} else {
		return "https://www.hodl.zone/" + safeUrl
	}
}

func (c App) Sandbox() revel.Result {
	return c.Render()
}

func (c App) Index() revel.Result {
	return c.RenderTemplate("App/Index.html")
}

func (c App) Landing() revel.Result {
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

	stringToken, err := cryption.NewJWTString(email, state.JWTSecret, cryption.JWT_EXPIRY_TIME)
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

	stringToken, err := cryption.NewJWTString(email, state.JWTSecret, cryption.JWT_EXPIRY_TIME)
	if err != nil {
		data[JSON_ERROR] = fmt.Sprintf("Unable to create JWT: %s", err.Error())
		c.Response.Status = 500
		return c.RenderJSON(data)
	}

	c.Session[cryption.COOKIE_JWT_MAP] = stringToken

	return c.RenderJSON(data)
}

func (c App) VerifyEmail() revel.Result {
	email := c.Params.Route.Get("email")
	hash := c.Params.Route.Get("hash")

	err := state.VerifyEmail(email, hash)
	if err != nil {
		fmt.Printf("WARNING: Attempt to verify email: %s hash: %s, error: %s\n", email, hash, err.Error())
		return c.NotFound("Invalid link. Please verify your email again.")
	}
	c.ViewArgs["email"] = email
	return c.RenderTemplate("App/verifiedEmailSuccess.html")
}

func (c App) NewPassRequestGET() revel.Result {
	c.ViewArgs["get"] = true
	return c.RenderTemplate("App/NewPassRequest.html")
}

func (c App) NewPassRequestPOST() revel.Result {
	e := c.Params.Form.Get("email")

	tokenString, err := state.GetNewJWTOTP(e)
	if err != nil {
		fmt.Printf("ERROR: getting new JWTOTP email: [%s] error:%s\n", err.Error())
		c.Response.Status = 500
		return c.RenderTemplate("errors/500.html")
	}

	emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_USER, []string{
		e,
	}, "Reset Password")

	err = emailRequest.ParseTemplate("newpassword.html", struct {
		Link string
	}{
		MakeURL("newpass/response/" + tokenString),
	})

	if err != nil {
		fmt.Printf("ERROR: Parsing template: %s\n", err.Error())
		c.Response.Status = 500
		return c.RenderTemplate("errors/500.html")
	}

	if err = emailRequest.SendEmail(); err != nil {
		fmt.Printf("ERROR: Sending new password email: %s\n", err.Error())
		c.Response.Status = 500
		return c.RenderTemplate("errors/500.html")
	}

	c.ViewArgs["get"] = false
	return c.RenderTemplate("App/NewPassRequest.html")
}

func (c App) NewPassResponseGet() revel.Result {
	c.ViewArgs["get"] = true
	c.ViewArgs["tokenString"] = c.Params.Route.Get("jwt")
	return c.RenderTemplate("App/NewPass.html")
}

func (c App) NewPassResponsePost() revel.Result {
	tokenString := c.Params.Route.Get("jwt")
	c.ViewArgs["get"] = false

	c.ViewArgs["success"] = true
	if !state.CompareClearJWTOTP(tokenString) {
		c.ViewArgs["success"] = false
		fmt.Printf("ERROR: with new pass request JWTOTP: %s\n", tokenString)
		c.Response.Status = 400
	}
	return c.RenderTemplate("App/NewPass.html")
}
