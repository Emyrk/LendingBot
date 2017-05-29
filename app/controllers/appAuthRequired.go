package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/app/core/cryption"
	"github.com/revel/revel"
)

type AppAuthRequired struct {
	*revel.Controller
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

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	_, err := cryption.VerifyJWT(tokenString, state.JWTSecret)
	if err != nil {
		fmt.Printf("WARNING: AuthUser failed JWT Token: [%s] and error: %s\n", tokenString, err.Error())
		return r.Redirect(App.Index)
	}
	return nil
}
