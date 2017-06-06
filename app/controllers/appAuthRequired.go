package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
)

var _ = userdb.SaltLength
var SkipAuth = false

type AppAuthRequired struct {
	*revel.Controller
	Email string
}

type PoloniexKeys struct {
	PoloniexKey    string `json:"poloniexkey"`
	PoloniexSecret string `json:"poloniexsecret"`
}

type Enable2fa struct {
	Pass   string `json:"pass"`
	Enable bool   `json:"enable"`
	Token  string `json:"token"`
}

type Pass struct {
	Pass string `json:"pass"`
}

func (r AppAuthRequired) unmarshalPoloniexKeys(body io.ReadCloser) *PoloniexKeys {
	var jsonPoloniexKeys PoloniexKeys
	err := json.NewDecoder(body).Decode(&jsonPoloniexKeys)
	if err != nil {
		fmt.Printf("Error unmarshaling json poloniex keys: %s\n", err.Error())
		return nil
	}
	defer body.Close()
	return &jsonPoloniexKeys
}

func (r AppAuthRequired) unmarshal2fa(body io.ReadCloser) *Enable2fa {
	var json2fa Enable2fa
	err := json.NewDecoder(body).Decode(&json2fa)
	if err != nil {
		fmt.Printf("Error unmarshaling 2fa: %s\n", err.Error())
		return nil
	}
	defer body.Close()
	return &json2fa
}

func (r AppAuthRequired) unmarshalPass(body io.ReadCloser) string {
	var jsonPass Pass
	err := json.NewDecoder(body).Decode(&jsonPass)
	if err != nil {
		fmt.Printf("Error unmarshaling pass: %s\n", err.Error())
		return ""
	}
	defer body.Close()
	return jsonPass.Pass
}

func (r AppAuthRequired) Dashboard() revel.Result {
	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Println("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}
	r.ViewArgs["UserLevel"] = fmt.Sprintf("%d", u.Level)
	r.ViewArgs["email"] = u.Username
	return r.Render()
}

func (r AppAuthRequired) Logout() revel.Result {
	DeleteCacheToken(r.Session.ID())
	return r.Redirect(App.Index)
}

func (r AppAuthRequired) InfoDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/InfoDashboard.html")
}

func (r AppAuthRequired) Enable2FA() revel.Result {
	data := make(map[string]interface{})
	json2fa := r.unmarshal2fa(r.Request.Body)
	if json2fa == nil {
		fmt.Printf("Error grabbing 2fa err\n")
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	email := r.Session[SESSION_EMAIL]

	err := state.Enable2FA(email, json2fa.Pass, json2fa.Token, json2fa.Enable)
	if err != nil {
		fmt.Printf("Error enabling 2fa err: %s\n", err.Error())
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	u, _ := state.FetchUser(email)
	data[JSON_DATA] = fmt.Sprintf("%t", u.Enabled2FA)

	return r.RenderJSON(data)
}

func (r AppAuthRequired) InfoAdvancedDashboard() revel.Result {
	return r.RenderTemplate("AppAuthRequired/InfoAdvancedDashboard.html")
}

func (r AppAuthRequired) SetPoloniexKeys() revel.Result {
	data := make(map[string]interface{})

	email := r.Session[SESSION_EMAIL]
	err := state.SetUserKeys(email, r.Params.Form.Get("poloniexkey"), r.Params.Form.Get("poloniexsecret"))
	if err != nil {
		fmt.Printf("Error authenticating setting Poloniex Keys err: %s\n", err.Error())
		data[JSON_ERROR] = "Error with Setting Poloniex Keys"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	poloniexKeys := &PoloniexKeys{
		r.Params.Form.Get("poloniexkey"),
		r.Params.Form.Get("poloniexsecret"),
	}

	poloniexKeys.PoloniexSecret = ""

	d, err := json.Marshal(poloniexKeys)
	if err != nil {
		fmt.Printf("Error marshalling poloniex keys, err: %s\n", err.Error())
		data[JSON_ERROR] = "Error marshalling keys"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	data[JSON_DATA] = fmt.Sprintf("%s", d)
	return r.RenderJSON(data)
}

func (r AppAuthRequired) SettingsDashboardUser() revel.Result {
	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	r.ViewArgs["verified"] = fmt.Sprintf("%t", u.Verified)
	r.ViewArgs["has2FA"] = fmt.Sprintf("%t", u.Has2FA)
	r.ViewArgs["enabled2FA"] = fmt.Sprintf("%t", u.Enabled2FA)

	if u.PoloniexKeys.APIKeyEmpty() {
		r.ViewArgs["poloniexKey"] = ""
	} else {
		s, err := u.PoloniexKeys.DecryptAPIKeyString(u.GetCipherKey(state.CipherKey))
		if err != nil {
			fmt.Printf("Error decrypting Api Keys String: %s\n", err.Error())
			s = ""
		}
		r.ViewArgs["poloniexKey"] = s
	}

	if u.PoloniexKeys.SecretKeyEmpty() {
		r.ViewArgs["poloniexSecret"] = ""
	} else {
		r.ViewArgs["poloniexSecret"] = ""
	}

	fmt.Println(r.ViewArgs["poloniexKey"], r.ViewArgs["poloniexSecret"])

	return r.RenderTemplate("AppAuthRequired/SettingsDashboardUser.html")
}

func (r AppAuthRequired) SettingsDashboardLending() revel.Result {
	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	if u.PoloniexKeys.APIKeyEmpty() {
		r.ViewArgs["poloniexKey"] = ""
	} else {
		s, err := u.PoloniexKeys.DecryptAPIKeyString(u.GetCipherKey(state.CipherKey))
		if err != nil {
			fmt.Printf("Error decrypting Api Keys String: %s\n", err.Error())
			s = ""
		}
		r.ViewArgs["poloniexKey"] = s
	}

	if u.PoloniexKeys.SecretKeyEmpty() {
		r.ViewArgs["poloniexSecret"] = ""
	} else {
		r.ViewArgs["poloniexSecret"] = ""
	}

	r.ViewArgs["lendingEnabled"] = u.PoloniexEnabled

	return r.RenderTemplate("AppAuthRequired/SettingsDashboardLending.html")
}

func (r AppAuthRequired) Create2FA() revel.Result {
	pass := r.unmarshalPass(r.Request.Body)

	data := make(map[string]interface{})

	qr, err := state.Add2FA(r.Session[SESSION_EMAIL], pass)
	if err != nil {
		fmt.Printf("Error authenticating 2fa err: %s\n", err.Error())
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	data[JSON_DATA] = qr

	return r.RenderJSON(data)
}

func (r AppAuthRequired) RequestEmailVerification() revel.Result {
	u, _ := state.FetchUser(r.Session[SESSION_EMAIL])

	data := make(map[string]interface{})

	if u.Verified {
		fmt.Printf("WARNING: User already verified: %s\n", r.Session[SESSION_EMAIL])
		data[JSON_ERROR] = "Bad Request"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	link := MakeURL("verifyemail/" + url.QueryEscape(u.Username) + "/" + url.QueryEscape(u.VerifyString))

	emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_USER, []string{
		r.Session[SESSION_EMAIL],
	}, "Verify Account")

	err := emailRequest.ParseTemplate("verify.html", struct {
		Link string
	}{
		link,
	})
	fmt.Printf("Template %s\n", emailRequest.Body)
	if err != nil {
		fmt.Printf("ERROR: Parsing template: %s\n", err)
		data[JSON_ERROR] = "Internal Error"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	if err = emailRequest.SendEmail(); err != nil {
		fmt.Printf("ERROR: Sending email: %s\n", err)
		data[JSON_ERROR] = "Internal Error"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	return r.RenderJSON(data)
}

func (r AppAuthRequired) EnableUserLending() revel.Result {
	data := make(map[string]interface{})
	err := state.EnableUserLending(r.Session[SESSION_EMAIL], r.Params.Form.Get("enable") == "true")
	if err != nil {
		fmt.Printf("WARNING: User failed to enable/disable lending: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Bad Request"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	return r.RenderJSON(data)
}

func (r AppAuthRequired) ChangePassword() revel.Result {
	data := make(map[string]interface{})

	err := state.SetUserNewPass(r.Session[SESSION_EMAIL], r.Params.Form.Get("pass"), r.Params.Form.Get("passnew"))
	if err != nil {
		fmt.Printf("WARNING: User failed to reset pass: [%s] error: %s\n", r.Session[SESSION_EMAIL], err.Error())
		data[JSON_ERROR] = "Bad Request"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	return r.RenderJSON(data)
}

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	if !ValidCacheEmail(r.Session.ID(), r.Session[SESSION_EMAIL]) {
		fmt.Printf("WARNING: AuthUser has invalid cache: [%s] sessionId:[%s]\n", r.Session[SESSION_EMAIL], r.Session.ID())
		if SkipAuth {
			return nil
		}
		r.Session[SESSION_EMAIL] = ""
		return r.Redirect(App.Index)
	}

	err := SetCacheEmail(r.Session.ID(), r.Session[SESSION_EMAIL])
	if err != nil {
		fmt.Printf("WARNING: AuthUser failed to set cache: [%s] and error: %s\n", r.Session.ID(), err.Error())
		r.Session[SESSION_EMAIL] = ""
		return r.Redirect(App.Index)
	}
	//do not cache auth pages
	r.Response.Out.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")

	return nil
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
