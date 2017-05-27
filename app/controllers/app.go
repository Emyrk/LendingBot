package controllers

import (
	"github.com/DistributedSolutions/LendingBot/database"
	"github.com/revel/revel"
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

	if database.Login(email, pass) {
		data["error"] = nil
		data["data"] = nil
		return c.RenderJSON(data)
	} else {
		data["error"] = "Invalid login"
		data["data"] = nil
		return c.RenderJSON(data)
	}
}
