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

	poloniexKeys := r.unmarshalPoloniexKeys(r.Request.Body)
	if poloniexKeys == nil {
		fmt.Println("Error unmarshalling poloniex keys")
		data[JSON_ERROR] = "Error with request"
		r.Response.Status = 400
		return r.RenderJSON(data)
	}
	email := r.Session[SESSION_EMAIL]

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
		r.ViewArgs["poloniexSecret"] = "********"
	}

	return r.RenderTemplate("AppAuthRequired/SettingsDashboard.html")
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

//called before any auth required function
func (r AppAuthRequired) AuthUser() revel.Result {
	r.Response.Out.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
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

// UserBalanceDetails is their current lending balances
type UserBalanceDetails struct {
	CurrencyMap map[string]float64
	Percent     map[string]float64
}

func newUserBalanceDetails() *UserBalanceDetails {
	u := new(UserBalanceDetails)
	u.CurrencyMap = make(map[string]float64)
	u.Percent = make(map[string]float64)
	return u
}

func (u *UserBalanceDetails) compute() {
	total := float64(0)
	for _, v := range u.CurrencyMap {
		total += v
	}

	for k, v := range u.CurrencyMap {
		u.Percent[k] = v / total
	}
}

// UserDashboard is the main page for users that have poloniex lending setup
func (r AppAuthRequired) UserDashboard() revel.Result {
	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Println("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}

	userStats, err := state.GetUserStatistics(email, 2)
	if err != nil {
		// HANDLE
	}

	balanceDetails := newUserBalanceDetails()
	today := newUserDashRow0()
	l := len(userStats)
	if l > 0 && len(userStats[0]) > 0 {
		now := userStats[0][0]
		// Set balance ratios
		balanceDetails.CurrencyMap = now.TotalCurrencyMap
		balanceDetails.compute()

		today.LoanRate = now.AverageActiveRate
		today.BTCLent = now.ActiveLentBalance
		today.BTCNotLent = now.AverageOnOrderRate + now.AvailableBalance
		today.LendingPercent = today.BTCLent / (today.BTCLent + today.BTCNotLent)

		yesterday := userdb.GetDayAvg(userStats[1])
		if yesterday != nil {
			today.LoanRateChange = percentChange(today.LoanRate, yesterday.LoanRate)
			today.BTCLentChange = percentChange(today.BTCLent, yesterday.BTCLent)
			today.BTCNotLentChange = percentChange(today.BTCNotLent, yesterday.BTCNotLent)
			today.LendingPercentChange = percentChange(today.LendingPercent, yesterday.LendingPercent)
		}
	}

	completeLoans, err := state.PoloniexAuthenticatedLendingHistory(u.Username, "", "")
	dataShort := completeLoans.Data
	if len(dataShort) > 5 {
		dataShort = dataShort[:5]
	}
	r.ViewArgs["CompleteLoans"] = completeLoans.Data
	r.ViewArgs["Today"] = today
	r.ViewArgs["Balances"] = balanceDetails
	return r.Render()
}

func abs(a float64) float64 {
	if a < 0 {
		return a * -1
	}
	return a
}

func percentChange(a float64, b float64) float64 {
	if a == 0 || b == 0 {
		return 0
	}
	change := ((a - b) / a) * 100
	if abs(change) < 0.001 {
		return 0
	}
	return change
}
