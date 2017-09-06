package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/revel/revel"

	// Init logger
	_ "github.com/Emyrk/LendingBot/src/log"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Println

var state *core.State
var appLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "app",
})

func init() {
	RegisterPrometheus()
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

func MakeURL(safeUrl string) string {
	if revel.DevMode {
		return "http://localhost:9000/" + safeUrl
	} else {
		return "https://www.hodl.zone/" + safeUrl
	}
}

func (c App) Sandbox() revel.Result {
	if !revel.DevMode {
		return c.RenderTemplate("errors/404.html")
	}
	c.ViewArgs["UserLevel"] = "1000"
	c.ViewArgs["Inverse"] = true
	return c.Render()
}

func (c App) Index() revel.Result {
	llog := appLog.WithField("method", "Index")
	stats, err := json.Marshal(Balancer.RateCalculator.GetQuickPoloniexStatisitics("BTC"))
	if err != nil {
		llog.Errorf("ERROR CRUCIAL!!!: retrieving index stats: %s", err.Error())
		c.ViewArgs["poloniexStats"] = "null"
	} else {
		c.ViewArgs["poloniexStats"] = string(stats)
	}
	AppPageHitIndex.Inc()
	c.ViewArgs["IsLoggedIn"] = len(c.Session[SESSION_EMAIL]) > 0
	return c.RenderTemplate("App/Index.html")
}

func (c App) FAQ() revel.Result {
	c.ViewArgs["Inverse"] = true
	AppPageHitFAQ.Inc()
	return c.RenderTemplate("App/FAQ.html")
}

func (c App) Donate() revel.Result {
	c.ViewArgs["Inverse"] = true
	AppPageHitDonate.Inc()
	return c.RenderTemplate("App/Donate.html")
}

func (c App) TermsAndPrivacy() revel.Result {
	c.ViewArgs["Inverse"] = true
	AppPageHitDonate.Inc()
	return c.RenderTemplate("App/TermsAndPrivacy.html")
}

func (c App) Information() revel.Result {
	c.ViewArgs["Inverse"] = true
	AppPageHitInformation.Inc()
	return c.RenderTemplate("App/Information.html")
}

func (c App) Contact() revel.Result {
	c.ViewArgs["Inverse"] = true
	AppPageHitContact.Inc()
	return c.RenderTemplate("App/Contact.html")
}

func (c App) Landing() revel.Result {
	AppPageHitLanding.Inc()
	return c.Render()
}

func (c App) Login() revel.Result {
	llog := appLog.WithField("method", "Login")

	email := c.Params.Form.Get("email")
	pass := c.Params.Form.Get("pass")
	twofa := c.Params.Form.Get("twofa")

	data := make(map[string]interface{})
	ok, _, err := state.AuthenticateUser2FA(email, pass, twofa)
	if err != nil {
		llog.Errorf("Error authenticating err: %s", err.Error())
		data[JSON_ERROR] = "Invalid username, password or 2fa, please try again."
		c.Response.Status = 500
		return c.RenderJSON(data)
	}
	if !ok {
		llog.Errorf("Error authenticating email: %s", email)
		data[JSON_ERROR] = "Invalid username, password or 2fa, please try again."
		c.Response.Status = 400
		return c.RenderJSON(data)
	}

	c.Session[SESSION_EMAIL] = email

	httpCookie, err := SetCacheEmail(c.Session.ID(), c.ClientIP, email)
	if err != nil {
		llog.Errorf("Error setting email cache: %s", err.Error())
	} else {
		c.SetCookie(httpCookie)
	}

	AppPageHitLogin.Inc()

	return c.RenderJSON(data)
}

func (c App) Register() revel.Result {
	llog := appLog.WithField("method", "Register")

	e := c.Params.Form.Get("email")
	pass := c.Params.Form.Get("pass")
	code := c.Params.Form.Get("code")

	data := make(map[string]interface{})

	if code != "" {
		ok, err := state.ReferralCodeExists(code)
		if err != nil {
			llog.Errorf("Error claiming invite code: %s", err.Error())
			data[JSON_ERROR] = "Invite code invalid."
			c.Response.Status = 500
			return c.RenderJSON(data)
		}
		if !ok {
			llog.Warnf("Invite code[%s] does not exist.", code)
			data[JSON_ERROR] = "Invite code invalid."
			c.Response.Status = 500
			return c.RenderJSON(data)
		}
	}

	// ok, err := state.ClaimInviteCode(e, code)
	// if err != nil {
	// 	llog.Errorf("Error claiming invite code: %s", err.Error())
	// 	data[JSON_ERROR] = "Invite code invalid."
	// 	c.Response.Status = 500
	// 	return c.RenderJSON(data)
	// }

	// if !ok {
	// 	llog.Warningf("Warning invite code invalid: %s", err.Error())
	// 	data[JSON_ERROR] = "Invite code invalid."
	// 	c.Response.Status = 400
	// 	return c.RenderJSON(data)
	// }

	apiErr := state.NewUser(e, pass)
	if apiErr != nil {
		llog.Errorf("Error registering user: %s", apiErr.LogError.Error())
		data[JSON_ERROR] = apiErr.UserError.Error()
		c.Response.Status = 400
		return c.RenderJSON(data)
	}

	apiErr = state.SetUserReferee(e, code)
	if apiErr != nil {
		llog.Errorf("Error setting referee code: %s", apiErr.LogError.Error())
	}

	c.Session[SESSION_EMAIL] = e

	u, err := state.FetchUser(e)
	if err != nil {
		llog.Errorf("Error fetching new user: %s", err)
	} else {
		httpCookie, err := SetCacheEmail(c.Session.ID(), c.ClientIP, u.Username)
		if err != nil {
			llog.Errorf("Error setting email cache: %s", err.Error())
		} else {
			c.SetCookie(httpCookie)
		}

		link := MakeURL("verifyemail/" + url.QueryEscape(u.Username) + "/" + url.QueryEscape(u.VerifyString))

		emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_NO_REPLY, []string{
			c.Session[SESSION_EMAIL],
		}, "Verify Account")

		err = emailRequest.ParseTemplate("verify.html", struct {
			Link string
		}{
			link,
		})

		if err != nil {
			llog.Errorf("Error register parsing template: %s", err)
		} else {
			if err = emailRequest.SendEmail(); err != nil {
				llog.Errorf("Error register sending email: %s", err)
			}
		}
	}

	AppPageHitRegister.Inc()

	return c.RenderJSON(data)
}

func (c App) VerifyEmail() revel.Result {
	llog := appLog.WithField("method", "VerifyEmail")

	email := c.Params.Route.Get("email")
	hash := c.Params.Route.Get("hash")

	err := state.VerifyEmail(email, hash)
	if err != nil {
		llog.Warningf("Attempt to verify email: %s hash: %s, error: %s", email, hash, err.Error())
		return c.NotFound("Invalid link. Please verify your email again.")
	}
	c.ViewArgs["email"] = email

	AppPageHitVerifyEmail.Inc()
	return c.RenderTemplate("App/verifiedEmailSuccess.html")
}

func (c App) NewPassRequestGET() revel.Result {
	c.ViewArgs["get"] = true
	c.ViewArgs["Inverse"] = true
	return c.RenderTemplate("App/NewPassRequest.html")
}

func (c App) PaymentNotification() revel.Result {
	llog := appLog.WithField("method", "PaymentNotification")

	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		llog.Errorf(err.Error())
		c.Response.Status = 500
		return c.RenderJSON(struct {
			Error string
			Spot  int
		}{
			err.Error(),
			1,
		})
	}

	err = CoinbaseWatcher.IncomingNotification(data)
	if err != nil {
		llog.Errorf(err.Error())
		c.Response.Status = 500
		return c.RenderJSON(struct {
			Error string
			Spot  int
		}{
			err.Error(),
			2,
		})
	}

	llog.Infof(string(data))

	return c.RenderJSON(struct {
		Success string
	}{
		"success",
	})
}

func (c App) NewPassRequestPOST() revel.Result {
	llog := appLog.WithField("method", "NewPassRequestPOST")

	e := c.Params.Form.Get("email")

	c.ViewArgs["get"] = false
	c.ViewArgs["Inverse"] = true
	AppPageHitNewPassGet.Inc()

	tokenString, err := state.GetNewJWTOTP(e)
	if err != nil {
		llog.Errorf("Error getting new JWTOTP email: [%s] error:%s", e, err.Error())
		// c.Response.Status = 500
		// return c.RenderTemplate("errors/500.html")
		return c.RenderTemplate("App/NewPassRequest.html")
	}

	emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_NO_REPLY, []string{
		e,
	}, "Reset Password")

	err = emailRequest.ParseTemplate("newpassword.html", struct {
		Link string
	}{
		MakeURL("newpass/response/" + tokenString),
	})

	if err != nil {
		llog.Errorf("Error parsing template: %s", err.Error())
		// c.Response.Status = 500
		// return c.RenderTemplate("errors/500.html")
		return c.RenderTemplate("App/NewPassRequest.html")
	}

	if err = emailRequest.SendEmail(); err != nil {
		llog.Errorf("Error sending new password email: %s", err.Error())
		// c.Response.Status = 500
		// return c.RenderTemplate("errors/500.html")
		return c.RenderTemplate("App/NewPassRequest.html")
	}

	return c.RenderTemplate("App/NewPassRequest.html")
}

func (c App) NewPassResponseGet() revel.Result {
	c.ViewArgs["get"] = true
	c.ViewArgs["tokenString"] = c.Params.Route.Get("jwt")
	c.ViewArgs["Inverse"] = true
	AppPageHitNewPassGet.Inc()
	return c.RenderTemplate("App/NewPass.html")
}

func (c App) NewPassResponsePost() revel.Result {
	llog := appLog.WithField("method", "NewPassResponsePost")

	tokenString := c.Params.Route.Get("jwt")
	pass := c.Params.Form.Get("pass")
	c.ViewArgs["get"] = false

	c.ViewArgs["success"] = true
	username, ok := state.SetNewPasswordJWTOTP(tokenString, pass)
	if !ok {
		c.ViewArgs["success"] = false
		llog.Errorf("Error with new pass request JWTOTP: %s", tokenString)
		c.Response.Status = 400
	}

	emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_NO_REPLY, []string{
		username,
	}, "Password Change")
	err := emailRequest.ParseTemplate("passwordchange.html", nil)
	if err != nil {
		llog.Errorf("Error parsing template for user[%s]: %s", username, err)
		return c.RenderError(&revel.Error{
			Title:       "500 Error.",
			Description: "Internal Error. Please contact support at: support@hodl.zone",
		})
	}
	if err = emailRequest.SendEmail(); err != nil {
		llog.Errorf("Error sending email for user[%s]: %s", username, err)
		return c.RenderError(&revel.Error{
			Title:       "500 Error.",
			Description: "Internal Error. Please contact support at: support@hodl.zone",
		})
	}

	c.ViewArgs["Inverse"] = true
	AppPageHitNewPassPost.Inc()
	return c.RenderTemplate("App/NewPass.html")
}

func (c App) ValidAuth() revel.Result {
	data := make(map[string]interface{})
	data[JSON_DATA] = len(c.Session[SESSION_EMAIL]) > 0
	return c.RenderJSON(data)
}

func (c App) LanguageChange() revel.Result {
	langCookie := &http.Cookie{
		Name:  revel.Config.StringDefault("cookie.prefix", "HODL") + "_LANG",
		Value: c.Params.Route.Get("l"),
		Path:  "/",
	}
	c.SetCookie(langCookie)

	return c.Redirect("/")
}

//called before any auth required function
func (c App) AppAuthUser() revel.Result {
	email := c.Session[SESSION_EMAIL]
	if email != "" && !ValidCacheEmail(c.Session.ID(), c.ClientIP, email) {
		c.Session[SESSION_EMAIL] = ""
	}

	return nil
}
