package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/userdb"
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
	if !ValidCacheEmail(s.Session.ID(), s.Session[SESSION_EMAIL]) {
		fmt.Printf("WARNING: AuthUser has invalid cache: [%s] sessionId:[%s]\n", s.Session[SESSION_EMAIL], s.Session.ID())
		s.Session[SESSION_EMAIL] = ""
		return s.Redirect(App.Index)
	}

	err := SetCacheEmail(s.Session.ID(), s.Session[SESSION_EMAIL])
	if err != nil {
		fmt.Printf("WARNING: AuthUser failed to set cache: [%s] and error: %s\n", s.Session.ID(), err.Error())
		s.Session[SESSION_EMAIL] = ""
		return s.Redirect(App.Index)
	}

	// tokenString := s.Session[cryption.COOKIE_JWT_MAP]
	// email, err := cryption.VerifyJWTGetEmail(tokenString, state.JWTSecret)
	// if err != nil {
	// 	fmt.Printf("WARNING: AuthUser SysAdmin failed JWT Token: %s\n", tokenString)
	// 	return s.Redirect(App.Index)
	// }

	u, err := state.FetchUser(s.Session[SESSION_EMAIL])
	if err != nil || u == nil {
		fmt.Printf("WARNING: AuthUser SysAdmin failed to fetch user: %s\n", s.Session[SESSION_EMAIL])
		return s.Redirect(App.Index)
	}

	if u.Level != userdb.SysAdmin {
		fmt.Printf("WARNING: IMPORTANT: AuthUser SysAdmin user level: %d trying to attempt access: %s\n", u.Level, s.Session[SESSION_EMAIL])
		return s.Redirect(App.Index)
	}

	return nil
}
