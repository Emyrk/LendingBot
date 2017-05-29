package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/app/core/cryption"
	"github.com/Emyrk/LendingBot/app/core/userdb"
	"github.com/revel/revel"
)

type AppSysAdmin struct {
	*revel.Controller
}

func (s AppSysAdmin) LogsDashboard() revel.Result {
	fmt.Println("logs")
	return s.RenderTemplate("AppSysAdmin/LogsDashboard.html")
}

func (s AppSysAdmin) ExportLogs() revel.Result {
	return s.Render()
}

func (s AppSysAdmin) DeleteLogs() revel.Result {
	return s.Render()
}

//called before any auth required function
func (s AppSysAdmin) AuthUserSysAdmin() revel.Result {
	tokenString := s.Session[cryption.COOKIE_JWT_MAP]
	email, err := cryption.VerifyJWT(tokenString, state.JWTSecret)
	if err != nil {
		fmt.Printf("WARNING: AuthUser SysAdmin failed JWT Token: %s\n", tokenString)
		return s.Redirect(App.Index)
	}

	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Printf("WARNING: AuthUser SysAdmin failed to fetch user: %s\n", tokenString)
		return s.Redirect(App.Index)
	}

	if u.Level != userdb.SysAdmin {
		fmt.Printf("WARNING: IMPORTANT: AuthUser SysAdmin user level: %d trying to attempt access: %s\n", u.Level, tokenString)
		return s.Redirect(App.Index)
	}

	return nil
}
