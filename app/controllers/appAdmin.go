package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/userdb"
	ourlog "github.com/Emyrk/LendingBot/src/log"
	"github.com/revel/revel"
	log "github.com/sirupsen/logrus"
)

var caLog = log.WithField("package", "appAdmin")

type AppAdmin struct {
	*revel.Controller
}

func (s AppAdmin) DashboardUsers() revel.Result {
	return s.RenderTemplate("AppAdmin/DashboardUsers.html")
}

func (s AppAdmin) DashboardQueuerStatus() revel.Result {
	s.ViewArgs["QueuerStatus"] = Balancer.GetLastReportString()
	return s.RenderTemplate("AppAdmin/DashboardQueuerStatus.html")
}

func (s AppAdmin) DashboardLogs() revel.Result {
	llog := caLog.WithField("method", "DashboardLogs")
	logs, err := ourlog.ReadLogs()
	if err != nil {
		logs = fmt.Sprintf("Error reading log: %s", err.Error())
		llog.Errorf("Error reading logs: %s\n", err.Error())
	}
	s.ViewArgs["LogFile"] = logs
	return s.RenderTemplate("AppAdmin/DashboardLogs.html")
}

func (s AppAdmin) GetLogs() revel.Result {
	llog := caLog.WithField("method", "GetLogs")
	logs, err := ourlog.ReadLogs()
	if err != nil {
		logs = fmt.Sprintf("Error reading log: %s", err.Error())
		llog.Errorf("Error reading logs: %s\n", err.Error())
	}
	data := make(map[string]interface{})

	data["log"] = logs

	return s.RenderJSON(data)
}

//called before any auth required function
func (s AppAdmin) AuthUserAdmin() revel.Result {
	if !ValidCacheEmail(s.Session.ID(), s.Session[SESSION_EMAIL]) {
		fmt.Printf("WARNING: AuthUserSysAdmin has invalid cache: [%s] sessionId:[%s]\n", s.Session[SESSION_EMAIL], s.Session.ID())
		s.Session[SESSION_EMAIL] = ""
		return s.Redirect(App.Index)
	}

	err := SetCacheEmail(s.Session.ID(), s.Session[SESSION_EMAIL])
	if err != nil {
		fmt.Printf("WARNING: AuthUserSysAdmin failed to set cache: [%s] and error: %s\n", s.Session.ID(), err.Error())
		s.Session[SESSION_EMAIL] = ""
		return s.Redirect(App.Index)
	}

	if !state.HasUserPrivilege(s.Session[SESSION_EMAIL], userdb.Admin) {
		return s.Redirect(App.Index)
	}

	//do not cache auth pages yet
	s.Response.Out.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")

	s.SetCookie(GetTimeoutCookie())

	return nil
}

func (s AppAdmin) GetUserStats() revel.Result {
	email := s.Params.Form.Get("email")
	stats, bals := getUserStats(email)
	data := make(map[string]interface{})

	// Scrub for NaNs
	stats.scrub()
	bals.scrub()

	data["Stats"] = stats
	data["Bals"] = bals
	return s.RenderJSON(data)
}

func (s AppAdmin) GetUsers() revel.Result {
	data, err := getUsers()
	if err != nil {
		s.Response.Status = 500
		return s.RenderJSON(err)
	}
	return s.RenderJSON(data)
}
