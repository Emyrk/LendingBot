package controllers

import (
	"strconv"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
	log "github.com/sirupsen/logrus"
)

var appSysAdminLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "appSysAdmin",
})

type AppSysAdmin struct {
	*revel.Controller
}

func (s AppSysAdmin) SysAdminDashboard() revel.Result {
	return s.RenderTemplate("AppSysAdmin/SysAdminDashboard.html")
}

func getUsers() (map[string]interface{}, error) {
	llog := appSysAdminLog.WithField("method", "getUsers")

	data := make(map[string]interface{})

	safeUsers, err := state.GetAllUsers()
	if err != nil {
		llog.Warningf("Warning failed to get all users: error: %s\n", err.Error())
		data[JSON_ERROR] = "Failed to get all users. Look at logs."
		return nil, err
	}

	allLevels := userdb.AllLevels
	lArr := make([]string, len(allLevels), len(allLevels))
	for i, e := range allLevels {
		lArr[i] = userdb.LevelToString(e)
	}
	data[JSON_DATA] = struct {
		Users  []core.SafeUser `json:"users"`
		Levels []string        `json:"lev"`
	}{
		*safeUsers,
		lArr,
	}

	return data, nil
}

func (s AppSysAdmin) GetUsers() revel.Result {
	data, err := getUsers()
	if err != nil {
		s.Response.Status = 500
		return s.RenderJSON(err)
	}
	return s.RenderJSON(data)
}

func (s AppSysAdmin) DeleteUser() revel.Result {
	data := make(map[string]interface{})
	//need pass and email
	data[JSON_ERROR] = "Error not setup."
	return s.RenderJSON(data)
}

func (s AppSysAdmin) DeleteInvite() revel.Result {
	llog := appSysAdminLog.WithField("method", "DeleteInvite")

	data := make(map[string]interface{})

	err := state.DeleteInvite(s.Params.Form.Get("rawc"))
	if err != nil {
		llog.Warningf("Warning user failed to delete invite: [%s] error: %s\n", s.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to change delete invite."
		s.Response.Status = 500
		return s.RenderJSON(data)
	}

	return s.Render(data)
}

func (s AppSysAdmin) MakeInvite() revel.Result {
	llog := appSysAdminLog.WithField("method", "MakeInvite")

	data := make(map[string]interface{})

	h, err := strconv.Atoi(s.Params.Form.Get("hr"))
	if err != nil {
		llog.Warningf("Warning failed to convert hours to int: [%s] error: %s\n", s.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to change hours to int."
		s.Response.Status = 500
		return s.RenderJSON(data)
	}

	c, err := strconv.Atoi(s.Params.Form.Get("cap"))
	if err != nil {
		llog.Warningf("Warning failed to convert capacity to int: [%s] error: %s\n", s.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to change capacity to int."
		s.Response.Status = 500
		return s.RenderJSON(data)
	}

	t := time.Now().Add(time.Duration(h) * time.Hour)
	err = state.AddInviteCode(s.Params.Form.Get("rawc"), c, t)
	if err != nil {
		llog.Warningf("Warning user failed to create invite: [%s] error: %s\n", s.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to create invite."
		s.Response.Status = 500
		return s.RenderJSON(data)
	}

	return s.Render(data)
}

func (s AppSysAdmin) GetInvites() revel.Result {
	llog := appSysAdminLog.WithField("method", "GetInvites")

	data := make(map[string]interface{})

	inviteCodes, err := state.ListInviteCodes()
	if err != nil {
		llog.Warningf("Warning user failed to get invite codes: [%s] error: %s\n", s.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Error failed to get invite codes."
		s.Response.Status = 500
		return s.RenderJSON(data)
	}

	data[JSON_DATA] = inviteCodes

	return s.RenderJSON(data)
}

func (s AppSysAdmin) ChangeUserPrivilege() revel.Result {
	llog := appSysAdminLog.WithField("method", "ChangeUserPrivilege")

	data := make(map[string]interface{})

	u, _ := state.FetchUser(s.Session[SESSION_EMAIL])
	if !u.AuthenticatePassword(s.Params.Form.Get("pass")) {
		llog.Errorf("Error user failed to validate pass: [%s]\n", s.Session[SESSION_EMAIL])
		data[JSON_ERROR] = "Failed to change user privelege. Invalid pass."
		s.Response.Status = 400
		return s.RenderJSON(data)
	}

	priv, err := state.UpdateUserPrivilege(s.Params.Form.Get("email"), s.Params.Form.Get("priv"))
	if err != nil {
		llog.Warningf("Warning user failed to update privelege: [%s] error: %s\n", s.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to change user privelege."
		s.Response.Status = 500
		return s.RenderJSON(data)
	}

	data[JSON_DATA] = priv

	return s.Render(data)
}

func (s AppSysAdmin) LogsDashboard() revel.Result {
	if !revel.DevMode {
		return s.RenderTemplate("errors/404.html")
	}
	return s.RenderTemplate("AppSysAdmin/LogsDashboard.html")
}

func (s AppSysAdmin) ExportLogs() revel.Result {
	if !revel.DevMode {
		return s.RenderTemplate("errors/404.html")
	}
	return s.Render()
}

func (s AppSysAdmin) DeleteLogs() revel.Result {
	if !revel.DevMode {
		return s.RenderTemplate("errors/404.html")
	}
	return s.Render()
}

//called before any auth required function
func (s AppSysAdmin) AuthUserSysAdmin() revel.Result {
	llog := appSysAdminLog.WithField("method", "AuthUserSysAdmin")

	if !ValidCacheEmail(s.Session.ID(), s.ClientIP, s.Session[SESSION_EMAIL]) {
		llog.Warningf("Warning has invalid cache: [%s] sessionId:[%s]", s.Session[SESSION_EMAIL], s.Session.ID())
		s.Session[SESSION_EMAIL] = ""
		s.Response.Status = 403
		return s.RenderTemplate("errors/403.html")
	}

	err := SetCacheEmail(s.Session.ID(), s.ClientIP, s.Session[SESSION_EMAIL])
	if err != nil {
		llog.Warningf("Warning failed to set cache: [%s] and error: %s", s.Session.ID(), err.Error())
		s.Session[SESSION_EMAIL] = ""
		s.Response.Status = 403
		return s.RenderTemplate("errors/403.html")
	}

	if !state.HasUserPrivilege(s.Session[SESSION_EMAIL], userdb.SysAdmin) {
		s.Response.Status = 403
		return s.RenderTemplate("errors/403.html")
	}

	//do not cache auth pages yet
	// s.Response.Out.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")

	s.SetCookie(GetTimeoutCookie())

	return nil
}
