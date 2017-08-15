package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
	log "github.com/sirupsen/logrus"
)

var _ = userdb.SaltLength
var SkipAuth = false

var ignoredRoutes = map[string]bool{"/logout": true, "/dashboard/getactivitylog": true}

var appAuthrequiredLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "appAuthrequiredLog",
})

type AppAuthRequired struct {
	*revel.Controller
	Email string
}

type ExchangeKeys struct {
	ExchangeKey    string `json:"exchangekey"`
	ExchangeSecret string `json:"exchangesecret"`
}

type Enable2fa struct {
	Pass   string `json:"pass"`
	Enable bool   `json:"enable"`
	Token  string `json:"token"`
}

type Pass struct {
	Pass string `json:"pass"`
}

//Deprecated should use form on front end to avoid
func (r AppAuthRequired) unmarshal2fa(body io.ReadCloser) *Enable2fa {
	llog := appAuthrequiredLog.WithField("method", "unmarshal2fa")

	var json2fa Enable2fa
	err := json.NewDecoder(body).Decode(&json2fa)
	if err != nil {
		llog.Errorf("Error unmarshaling 2fa: %s\n", err.Error())
		return nil
	}
	defer body.Close()
	return &json2fa
}

//Deprecated should use form on front end to avoid
func (r AppAuthRequired) unmarshalPass(body io.ReadCloser) string {
	llog := appAuthrequiredLog.WithField("method", "unmarshalPass")

	var jsonPass Pass
	err := json.NewDecoder(body).Decode(&jsonPass)
	if err != nil {
		llog.Errorf("Error unmarshaling pass: %s\n", err.Error())
		return ""
	}
	defer body.Close()
	return jsonPass.Pass
}

func (r AppAuthRequired) Dashboard() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "Dashboard")

	r.ViewArgs["Version"] = VersionNumber

	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		llog.Errorf("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}
	r.ViewArgs["UserLevel"] = fmt.Sprintf("%d", u.Level)
	r.ViewArgs["email"] = u.Username
	r.ViewArgs["AvailableCoins"] = userdb.AvaiableCoins
	AppPageHitDashboard.Inc()
	return r.Render()
}

func (r AppAuthRequired) Logout() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "Logout")
	if err := DeleteCacheToken(r.Session.ID(), r.ClientIP, r.Session[SESSION_EMAIL]); err != nil {
		llog.Error("Error logging user[%s] out: %s", r.Session[SESSION_EMAIL], err.Error())
		r.Response.Status = 500
	}
	delete(r.Session, SESSION_EMAIL)
	AppPageHitInfoLogout.Inc()
	return r.Redirect(App.Index)
}

func (r AppAuthRequired) CoinDashboard() revel.Result {
	AppPageHitInfoDashboard.Inc()
	return r.RenderTemplate("AppAuthRequired/CoinDashboard.html")
}

func (r AppAuthRequired) InfoDashboard() revel.Result {
	AppPageHitInfoDashboard.Inc()
	return r.RenderTemplate("AppAuthRequired/InfoDashboard.html")
}

func (r AppAuthRequired) ChangeExpiry() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "ChangeExpiry")

	data := make(map[string]interface{})

	sesExp, err := strconv.Atoi(r.Params.Form.Get("sesexp"))
	if err != nil {
		llog.Errorf("Error parsing int user[%s] expiration: %s", r.Session[SESSION_EMAIL], r.Params.Form.Get("sesexp"))
		data[JSON_ERROR] = "Internal error. Please contact: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	//TODO: FIX test for setting minimum
	//	add in test for to long
	if time.Duration(sesExp)*time.Millisecond > CACHE_TIME_USER_SESSION_MAX {
		llog.Errorf("Error user[%s] attempting to set expiry larger than max: %d", r.Session[SESSION_EMAIL], time.Duration(sesExp)*time.Millisecond)
		data[JSON_ERROR] = "Session time to large."
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	// } else if time.Duration(sesExp)*time.Millisecond < CACHE_TIME_USER_SESSION_MIN {
	// 	llog.Errorf("Error user[%s] attempting to set expiry smaller than max: %d", r.Session[SESSION_EMAIL], time.Duration(sesExp)*time.Millisecond)
	// 	data[JSON_ERROR] = "Session time to small."
	// 	r.Response.Status = 500
	// 	return r.RenderJSON(data)
	// }

	err = state.SetUserExpiry(r.Session[SESSION_EMAIL], time.Duration(sesExp)*time.Millisecond)
	if err != nil {
		llog.Errorf("Error setting user[%s] exp: %s", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Internal error. Please contact: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	err = SetCacheDurEnd(r.Session[SESSION_EMAIL], time.Duration(sesExp)*time.Millisecond)
	if err != nil {
		llog.Errorf("Error setting user[%s] cache session exp: %s", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Internal error. Please contact: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	r.SetCookie(GetTimeoutCookie(time.Duration(sesExp) * time.Millisecond))
	return r.RenderJSON(data)
}

func (r AppAuthRequired) GetExpiry() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "GetExpiry")

	data := make(map[string]interface{})

	dur, err := GetCacheDur(r.Session[SESSION_EMAIL])
	if err != nil {
		llog.Errorf("Error getting user[%s] exp: %s", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Internal error. Please contact: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	data["sesexp"] = *dur / time.Millisecond
	return r.RenderJSON(data)
}

func (r AppAuthRequired) Enable2FA() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "Enable2FA")

	data := make(map[string]interface{})
	json2fa := r.unmarshal2fa(r.Request.Body)
	if json2fa == nil {
		llog.Errorf("Error grabbing 2fa err\n")
		data[JSON_ERROR] = "Internal error. Please contact: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	email := r.Session[SESSION_EMAIL]

	err := state.Enable2FA(email, json2fa.Pass, json2fa.Token, json2fa.Enable)
	if err != nil {
		llog.Errorf("Error enabling 2fa err: %s\n", err.Error())
		data[JSON_ERROR] = "Invalid password, please try again."
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	u, _ := state.FetchUser(email)
	data[JSON_DATA] = fmt.Sprintf("%t", u.Enabled2FA)

	AppPageHitInfoEnable2fa.Inc()
	return r.RenderJSON(data)
}

func (r AppAuthRequired) SetExchangeKeys() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "SetExchangeKeys")

	data := make(map[string]interface{})

	email := r.Session[SESSION_EMAIL]
	err := state.SetUserKeys(email, r.Params.Form.Get("exchangekey"), r.Params.Form.Get("exchangesecret"), userdb.UserExchange(r.Params.Form.Get("exch")))
	if err != nil {
		llog.Errorf("Error authenticating setting Poloniex Keys err: %s\n", err.Error())
		data[JSON_ERROR] = fmt.Sprintf("Error: %s", err.Error())
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	if r.Params.Form.Get("exch") == "bit" {
		Balancer.UpdateUserKey(email, balancer.BitfinexExchange)
	} else {
		Balancer.UpdateUserKey(email, balancer.PoloniexExchange)
	}

	poloniexKeys := &ExchangeKeys{
		r.Params.Form.Get("exchangekey"),
		r.Params.Form.Get("exchangesecret"),
	}

	poloniexKeys.ExchangeSecret = ""

	d, err := json.Marshal(poloniexKeys)
	if err != nil {
		llog.Errorf("Error marshalling poloniex keys, err: %s\n", err.Error())
		data[JSON_ERROR] = "Error marshalling keys"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	AppPageHitSetPoloKeys.Inc()
	data[JSON_DATA] = fmt.Sprintf("%s", d)
	return r.RenderJSON(data)
}

func (r AppAuthRequired) SettingsDashboardUser() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "SettingsDashboardUser")

	u, err := state.FetchUser(r.Session[SESSION_EMAIL])
	if err != nil {
		r.Response.Status = 500
		return r.RenderError(&revel.Error{
			Title:       "500 Error.",
			Description: "Looks like you are lost.",
		})
	}

	r.ViewArgs["verified"] = fmt.Sprintf("%t", u.Verified)
	r.ViewArgs["has2FA"] = fmt.Sprintf("%t", u.Has2FA)
	r.ViewArgs["enabled2FA"] = fmt.Sprintf("%t", u.Enabled2FA)
	r.ViewArgs["minSessionTime"] = fmt.Sprintf("%d", CACHE_TIME_USER_SESSION_MIN/time.Minute)
	r.ViewArgs["maxSessionTime"] = fmt.Sprintf("%d", CACHE_TIME_USER_SESSION_MAX/time.Hour*60)
	r.ViewArgs["currentSessionTime"] = fmt.Sprintf("%d", u.SessionExpiryTime/time.Minute)

	uss, err := GetUserActiveSessions(r.Session[SESSION_EMAIL], r.Session.ID())
	if err != nil {
		llog.Error("Error getting user active sessions: %s", err.Error())
	}
	b, err := json.Marshal(uss)
	if err != nil {
		llog.Errorf("Error marshalling user sessions: %s", err.Error())
		b = []byte("[]")
	}
	if len(uss) == 0 {
		b = []byte("[]")
	}
	r.ViewArgs["sessions"] = string(b)

	AppPageHitSetSettingDashUser.Inc()
	return r.RenderTemplate("AppAuthRequired/SettingsDashboardUser.html")
}

func (r AppAuthRequired) SettingsDashboardLending() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "SettingsDashboardLending")

	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	if u.PoloniexKeys.APIKeyEmpty() {
		r.ViewArgs["poloniexKey"] = ""
	} else {
		s, err := u.PoloniexKeys.DecryptAPIKeyString(u.GetCipherKey(state.CipherKey))
		if err != nil {
			llog.Errorf("Error decrypting Api Keys String: %s\n", err.Error())
			s = ""
		}
		r.ViewArgs["poloniexKey"] = s
	}

	if u.PoloniexKeys.SecretKeyEmpty() {
		r.ViewArgs["poloniexSecret"] = ""
	} else {
		r.ViewArgs["poloniexSecret"] = ""
	}

	AppPageHitSetSettingDashLend.Inc()
	return r.RenderTemplate("AppAuthRequired/SettingsDashboardLending.html")
}

func (r AppAuthRequired) Create2FA() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "Create2FA")

	pass := r.unmarshalPass(r.Request.Body)

	data := make(map[string]interface{})

	qr, err := state.Add2FA(r.Session[SESSION_EMAIL], pass)
	if err != nil {
		llog.Errorf("Error authenticating 2fa err: %s\n", err.Error())
		data[JSON_ERROR] = "Invalid password, please try again."
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	data[JSON_DATA] = qr

	AppPageHitCreate2fa.Inc()
	return r.RenderJSON(data)
}

func (r AppAuthRequired) SetReferee() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "SetReferee")
	data := make(map[string]interface{})
	err := state.SetUserReferee(r.Session[SESSION_EMAIL], r.Params.Form.Get("ref"))
	if err != nil {
		llog.Errorf("Error setting referee code: %s", err.LogError.Error())
		data[JSON_ERROR] = err.UserError.Error()
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	return r.RenderJSON(data)
}

func (r AppAuthRequired) PaymentDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/PaymentDashboard.html")
}

func (r AppAuthRequired) DespositDashboard() revel.Result {
	fmt.Println("HITDEPSOIT")
	return r.RenderTemplate("AppAuthRequired/DepositDashboard.html")
}

func (r AppAuthRequired) PredictionDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/PredictionDashboard.html")
}

func (r AppAuthRequired) RequestEmailVerification() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "RequestEmailVerification")

	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	data := make(map[string]interface{})

	if u.Verified {
		llog.Warningf("WARNING: User already verified: %s", r.Session[SESSION_EMAIL])
		data[JSON_ERROR] = "User already verified. No email sent."
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	link := MakeURL("verifyemail/" + url.QueryEscape(u.Username) + "/" + url.QueryEscape(u.VerifyString))

	emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_NO_REPLY, []string{
		r.Session[SESSION_EMAIL],
	}, "Verify Account")

	err := emailRequest.ParseTemplate("verify.html", struct {
		Link string
	}{
		link,
	})

	if err != nil {
		llog.Errorf("Error parsing template: %s\n", err)
		data[JSON_ERROR] = "Internal Error. Please contact support at: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	if err = emailRequest.SendEmail(); err != nil {
		llog.Errorf("Error sending email: %s\n", err)
		data[JSON_ERROR] = "Internal Error. Please contact support at: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	AppPageHitReqEmailVerify.Inc()
	return r.RenderJSON(data)
}

func (r AppAuthRequired) GetEnableUserLending() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "GetEnableUserLending")

	data := make(map[string]interface{})

	u, err := state.FetchUser(r.Session[SESSION_EMAIL])
	if err != nil {
		llog.Warningf("Warning user failed to find user: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Bad Request. Contact support@hodl.zone"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	var enableInterface interface{}
	var minLendInterface interface{}
	switch userdb.UserExchange(r.Params.Query.Get("exch")) {
	case userdb.PoloniexExchange:
		enableInterface = u.PoloniexEnabled
		minLendInterface = u.PoloniexMiniumLend
		break
	case userdb.BitfinexExchange:
		enableInterface = u.BitfinexEnabled
		minLendInterface = u.BitfinexMiniumumLend
		break
	default:
		llog.Warningf("Warning failed to set user exchange key: [%s] unknown exchange: %s\n", r.Session[SESSION_EMAIL], r.Params.Form.Get("exch"))
		data[JSON_ERROR] = "Bad Request. Exchange unknown. Contact support@hodl.zone"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	data[JSON_DATA] = struct {
		Enable interface{} `json:"enable"`
		Min    interface{} `json:"min"`
	}{
		enableInterface,
		minLendInterface,
	}

	AppPageHitEnableUserLending.Inc()
	return r.RenderJSON(data)
}

func (r AppAuthRequired) SetEnableUserLending() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "SetEnableUserLending")

	data := make(map[string]interface{})

	// /ENABLE USER LENDING
	err := state.EnableUserLending(r.Session[SESSION_EMAIL], r.Params.Form.Get("enable"), userdb.UserExchange(r.Params.Form.Get("exch")))
	if err != nil {
		llog.Warningf("Warning user failed to enable/disable lending: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to set and failed to enable/disable."
		r.Response.Status = 400
		return r.RenderJSON(data)
	}
	// /ENABLE USER LENDING

	// SET USER LENDING AMOUNT
	err = state.SetAllUserMinimumLoan(r.Session[SESSION_EMAIL], r.Params.Form.Get("min"), userdb.UserExchange(r.Params.Form.Get("exch")))
	if err != nil {
		llog.Warningf("Warning user failed set lending: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Failed to set values, but enable/disable went through."
		r.Response.Status = 400
		return r.RenderJSON(data)
	}
	// /SET USER LENDING AMOUNT

	AppPageHitSetEnableUserLending.Inc()
	return r.RenderJSON(data)
}

func (r AppAuthRequired) ChangePassword() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "ChangePassword")

	data := make(map[string]interface{})

	err := state.SetUserNewPass(r.Session[SESSION_EMAIL], r.Params.Form.Get("pass"), r.Params.Form.Get("passnew"))
	if err != nil {
		llog.Warningf("Warning user failed to reset pass: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Password incorrect. Please try again."
		r.Response.Status = 400
		return r.RenderJSON(data)
	}
	AppPageHitChangePass.Inc()
	return r.RenderJSON(data)
}

func (r AppAuthRequired) GetActivityLogs() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "GetActivityLogs")

	data := make(map[string]interface{})

	logs, err := state.GetActivityLog(r.Session[SESSION_EMAIL], r.Params.Query.Get("time"))
	if err != nil {
		llog.Warningf("Warning failed to get user activity logs: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Server error, failed to retrieve logs. Contact support: support@hodl.zone."
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	data["logs"] = logs
	return r.RenderJSON(data)
}

func (r AppAuthRequired) HasReferee() revel.Result {
	data := make(map[string]interface{})

	data["ref"] = state.HasSetReferee(r.Session[SESSION_EMAIL])

	return r.RenderJSON(data)
}

func (r AppAuthRequired) DeleteSession() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "DeleteSession")

	data := make(map[string]interface{})
	//delete session
	if err := DeleteCacheToken(r.Params.Form.Get("sesid"), r.ClientIP, r.Session[SESSION_EMAIL]); err != nil {
		llog.Error("Error deleting user session: %s", err.Error())
		data[JSON_ERROR] = "Server error, failed to delete session. Contact support: support@hodl.zone."
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	//get active sessions
	uss, err := GetUserActiveSessions(r.Session[SESSION_EMAIL], r.Session.ID())
	if err != nil {
		llog.Error("Error getting user active sessions after delete: %s", err.Error())
		data[JSON_ERROR] = "Server error, failed to delete session. Contact support: support@hodl.zone."
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	data["ses"] = uss

	return r.RenderJSON(data)
}

func (r AppAuthRequired) UserDashboard() revel.Result {
	if revel.DevMode || strings.Contains(revel.RunMode, "dev") {
		return r.RenderError(&revel.Error{
			Title:       "404 Error.",
			Description: "Looks like you are lost.",
		})
	}
	return r.Render()
}

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "AuthUser")

	if !ValidCacheEmail(r.Session.ID(), r.ClientIP, r.Session[SESSION_EMAIL]) {
		llog.Warningf("Warning invalid cache: email[%s] sessionId:[%s] url[%s]", r.Session[SESSION_EMAIL], r.Session.ID(), r.Request.URL)
		r.Session[SESSION_EMAIL] = ""
		r.Response.Status = 403
		return r.RenderTemplate("errors/403.html")
	}

	//must add rep
	if ignoredRoutes[r.Request.RequestURI] == true {
		return nil
	}

	AppPageAuthUser.Inc()

	httpCookie, err := SetCacheEmail(r.Session.ID(), r.ClientIP, r.Session[SESSION_EMAIL])
	if err != nil {
		llog.Warningf("Warning failed to set cache: email[%s] sessionId:[%s] url[%s] and error: %s", r.Session[SESSION_EMAIL], r.Session.ID(), r.Request.URL, err.Error())
		r.Session[SESSION_EMAIL] = ""
		r.Response.Status = 403
		return r.RenderTemplate("errors/403.html")
	} else {
		r.SetCookie(httpCookie)
	}

	//do not cache auth pages
	// r.Response.Out.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")

	return nil
}
