package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
	log "github.com/sirupsen/logrus"
)

var _ = userdb.SaltLength
var SkipAuth = false

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
	DeleteCacheToken(r.Session.ID())
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

	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	r.ViewArgs["verified"] = fmt.Sprintf("%t", u.Verified)
	r.ViewArgs["has2FA"] = fmt.Sprintf("%t", u.Has2FA)
	r.ViewArgs["enabled2FA"] = fmt.Sprintf("%t", u.Enabled2FA)

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

func (r AppAuthRequired) GenerateReferralCode() revel.Result {
	data := make(map[string]interface{})

	referralCode, err := state.GenerateUserReferralCode(r.Session[SESSION_EMAIL])
	if err != nil {
		llog.Errorf("Error generating referral code: %s\n", err.Error())
		data[JSON_ERROR] = "Internal Error. Please contact support at: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	data["referralcode"] = referralCode
	return r.RenderJSON(data)
}

func (r AppAuthRequired) SetReferee() revel.Result {
	data := make(map[string]interface{})
	err := state.SetUserReferee(r.Session[SESSION_EMAIL], r.Params.Form.Get("refereecode"))
	if err != nil {
		llog.Errorf("Error setting referee code: %s\n", err.Error())
		data[JSON_ERROR] = "Internal Error. Please contact support at: support@hodl.zone"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	return r.RenderJSON(data)
}

func (r AppAuthRequired) RequestEmailVerification() revel.Result {
	llog := appAuthrequiredLog.WithField("method", "RequestEmailVerification")

	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	data := make(map[string]interface{})

	if u.Verified {
		llog.Warningf("WARNING: User already verified: %s\n", r.Session[SESSION_EMAIL])
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

func (r AppAuthRequired) UserDashboard() revel.Result {
	if revel.DevMode {
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

	if !ValidCacheEmail(r.Session.ID(), r.Session[SESSION_EMAIL]) {
		llog.Warningf("Warning invalid cache: [%s] sessionId:[%s]\n", r.Session[SESSION_EMAIL], r.Session.ID())
		r.Session[SESSION_EMAIL] = ""
		r.Response.Status = 403
		return r.RenderTemplate("errors/403.html")
	}

	err := SetCacheEmail(r.Session.ID(), r.Session[SESSION_EMAIL])
	if err != nil {
		llog.Warningf("Warning failed to set cache: [%s] and error: %s\n", r.Session.ID(), err.Error())
		r.Session[SESSION_EMAIL] = ""
		r.Response.Status = 403
		return r.RenderTemplate("errors/403.html")
	}
	//do not cache auth pages
	// r.Response.Out.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")

	r.SetCookie(GetTimeoutCookie())

	AppPageAuthUser.Inc()
	return nil
}
