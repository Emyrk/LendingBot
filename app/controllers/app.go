package controllers

import (
	"fmt"
	"github.com/DistributedSolutions/LendingBot/app/core"
	"github.com/DistributedSolutions/LendingBot/app/core/cryption"
	"github.com/revel/revel"
	"net/http"
)

var state *core.State

func init() {
	state = core.NewState()
}

const (
	JSON_DATA  = "data"
	JSON_ERROR = "error"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Login() revel.Result {
	email := c.Params.Route.Get("email")
	pass := c.Params.Route.Get("pass")

	data := make(map[string]interface{})

	ok, _, err := state.AuthenticateUser(email, pass)
	if !ok {
		fmt.Printf("Error authenticating user: %s\n", err)
		data[JSON_ERROR] = "Invalid login"
		return c.RenderJSON(data)
	}

	stringToken, err := cryption.NewJWT(email, state.JWTSecret, cryption.JWT_EXPIRY_TIME)
	if err != nil {
		data[JSON_ERROR] = "Unable to create JWT"
		c.Response.Status = 500
		return c.RenderJSON(data)
	}

	jwt_cookie := &http.Cookie{Name: cryption.COOKIE_JWT_MAP, Value: stringToken}
	c.SetCookie(jwt_cookie)

	fmt.Printf("email: %s, pass: %s, cookie: %s\n", email, pass, stringToken)
	return c.RenderJSON(data)
}

func (c App) Register() revel.Result {
	email := c.Params.Route.Get("email")
	pass := c.Params.Route.Get("pass")

	data := make(map[string]interface{})

	err := state.NewUser(email, pass)
	if err != nil {
		data[JSON_ERROR] = "Unable to create new user"
		c.Response.Status = 500
		return c.RenderJSON(data)
	}

	stringToken, err := cryption.NewJWT(email, state.JWTSecret, cryption.JWT_EXPIRY_TIME)
	if err != nil {
		data[JSON_ERROR] = "Unable to create JWT"
		c.Response.Status = 500
		return c.RenderJSON(data)
	}

	jwt_cookie := &http.Cookie{Name: cryption.COOKIE_JWT_MAP, Value: stringToken}
	c.SetCookie(jwt_cookie)

	fmt.Printf("email: %s, pass: %s, cookie: %s\n", email, pass, stringToken)
	return c.RenderJSON(data)
}
