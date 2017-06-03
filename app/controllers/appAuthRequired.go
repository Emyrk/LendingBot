package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/Emyrk/LendingBot/src/core/cryption"
	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
)

var SkipAuth = false

type AppAuthRequired struct {
	*revel.Controller
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
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	email, _ := cryption.VerifyJWTGetEmail(tokenString, state.JWTSecret)
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

func (r AppAuthRequired) Enable2FA() revel.Result {
	data := make(map[string]interface{})
	json2fa := r.unmarshal2fa(r.Request.Body)
	if json2fa == nil {
		fmt.Printf("Error grabbing 2fa err\n")
		data[JSON_ERROR] = "Error with 2fa"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}
	email, _ := cryption.VerifyJWTGetEmail(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)

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

	poloniexKeys := r.unmarshalPoloniexKeys(r.Request.Body)
	if poloniexKeys == nil {
		fmt.Println("Error unmarshalling poloniex keys")
		data[JSON_ERROR] = "Error with request"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}
	email, _ := cryption.VerifyJWTGetEmail(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)

	err := state.SetUserKeys(email, poloniexKeys.PoloniexKey, poloniexKeys.PoloniexSecret)
	if err != nil {
		fmt.Printf("Error authenticating setting Poloniex Keys err: %s\n", err.Error())
		data[JSON_ERROR] = "Error with Setting Poloniex Keys"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	poloniexKeys.PoloniexSecret = "********"

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

func (r AppAuthRequired) SettingsDashboard() revel.Result {
	email, _ := cryption.VerifyJWTGetEmail(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)
	u, _ := state.FetchUser(email)

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
		r.ViewArgs["poloniexSecret"] = "********"
	}

	return r.RenderTemplate("AppAuthRequired/SettingsDashboard.html")
}

func (r AppAuthRequired) Create2FA() revel.Result {
	pass := r.unmarshalPass(r.Request.Body)
	email, _ := cryption.VerifyJWTGetEmail(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)

	data := make(map[string]interface{})

	qr, err := state.Add2FA(email, pass)
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
	e, _ := cryption.VerifyJWTGetEmail(r.Session[cryption.COOKIE_JWT_MAP], state.JWTSecret)
	u, _ := state.FetchUser(e)

	data := make(map[string]interface{})

	if u.Verified {
		fmt.Printf("WARNING: User already verified: %s\n", e)
		data[JSON_ERROR] = "Bad Request"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}

	link := MakeURL("verifyemail/" + url.QueryEscape(u.Username) + "/" + url.QueryEscape(u.VerifyString))

	emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_USER, []string{
		e,
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

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	if SkipAuth {
		return nil
	}
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	email, err := cryption.VerifyJWTGetEmail(tokenString, state.JWTSecret)
	if err != nil {
		fmt.Printf("WARNING: AuthUser failed JWT Token: [%s] and error: %s\n", tokenString, err.Error())
		return r.Redirect(App.Index)
	}

	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Printf("WARNING: AuthUser failed to fetch user: %s\n", tokenString)
		return r.Redirect(App.Index)
	}
	fmt.Println("FINSHED WIHT AUTH USER")

	return nil
}

// Struct to UserDash
type UserDashStructure struct {
}

type UserDashRow0 struct {
	LoanRate       float64
	BTCLent        float64
	BTCNotLent     float64
	LendingPercent float64

	LoanRateChange       float64
	BTCLentChange        float64
	BTCNotLentChange     float64
	LendingPercentChange float64

	// From poloniex call
	BTCEarned float64
}

func newUserDashRow0() *UserDashRow0 {
	r := new(UserDashRow0)
	r.LoanRate = 0
	r.BTCLent = 0
	r.BTCNotLent = 0
	r.LendingPercent = 0
	r.BTCEarned = 0

	r.LoanRateChange = 0
	r.BTCLentChange = 0
	r.BTCNotLentChange = 0
	r.LendingPercentChange = 0

	return r
}

/*
type UserStatistic struct {
	Username           string    `json:"username"`
	AvailableBalance   float64   `json:"availbal"`
	ActiveLentBalance  float64   `json:"availlent"`
	OnOrderBalance     float64   `json:"onorder"`
	AverageActiveRate  float64   `json:"activerate"`
	AverageOnOrderRate float64   `json:"onorderrate"`
	Time               time.Time `json:"time"`
	Currency           string    `json:"currency"`

	day int
}
*/

// UserDashboard is the main page for users that have poloniex lending setup
func (r AppAuthRequired) UserDashboard() revel.Result {
	tokenString := r.Session[cryption.COOKIE_JWT_MAP]
	email, _ := cryption.VerifyJWTGetEmail(tokenString, state.JWTSecret)
	u, err := state.FetchUser(email)

	if err != nil || u == nil {
		fmt.Println("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}

	userStats, err := state.GetUserStatistics(u.Username, 2)
	if err != nil {
		// HANDLE
	}

	today := newUserDashRow0()
	l := len(userStats)
	if l > 0 && len(userStats[0]) > 0 {
		now := userStats[0][0]
		today.LoanRate = now.AverageActiveRate
		today.BTCLent = now.ActiveLentBalance
		today.BTCNotLent = now.AverageOnOrderRate + now.AvailableBalance
		today.LendingPercent = today.BTCLent / (today.BTCLent + today.BTCNotLent)

		yesterday := userdb.GetDayAvg(userStats[1])
		if yesterday != nil {
			today.LoanRateChange = percentChange(yesterday.LoanRate, today.LoanRate)
			today.BTCLentChange = percentChange(yesterday.BTCLent, today.BTCLent)
			today.BTCNotLentChange = percentChange(yesterday.BTCNotLent, today.BTCNotLent)
			today.LendingPercentChange = percentChange(yesterday.LendingPercent, today.LendingPercent)
		}
	}

	r.ViewArgs["Today"] = today
	return r.Render()
}

func percentChange(a float64, b float64) float64 {
	return ((a - b) / a) * 100
}
