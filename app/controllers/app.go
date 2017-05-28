package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/DistributedSolutions/LendingBot/app/core"
	"github.com/DistributedSolutions/LendingBot/app/core/cryption"
	"github.com/revel/revel"
	"io"
)

var state *core.State

func init() {
	state = core.NewState()
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
	return c.Render()
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
	return r.Render()
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
