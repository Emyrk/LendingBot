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

func (r AppAuthRequired) CurrentUserStats() revel.Result {
	data := make(map[string]interface{})
	// json2fa := r.unmarshal2fa(r.Request.Body)
	// if json2fa == nil {
	// 	fmt.Printf("Error grabbing 2fa err\n")
	// 	data[JSON_ERROR] = "Error with 2fa"
	// 	r.Response.Status = 500
	// 	return r.RenderJSON(data)
	// }
	// email := r.Session[SESSION_EMAIL]

	// err := state.Enable2FA(email, json2fa.Pass, json2fa.Token, json2fa.Enable)
	// if err != nil {
	// 	fmt.Printf("Error enabling 2fa err: %s\n", err.Error())
	// 	data[JSON_ERROR] = "Error with 2fa"
	// 	r.Response.Status = 400
	// 	return r.RenderJSON(data)
	// }

	// u, _ := state.FetchUser(email)
	// data[JSON_DATA] = fmt.Sprintf("%t", u.Enabled2FA)

	return r.RenderJSON(data)
}

func (r AppAuthRequired) CurrentBalanceBreakdown() revel.Result {
	data := make(map[string]interface{})
	return r.RenderJSON(data)
}

func (r AppAuthRequired) LendingHistory() revel.Result {
	data := make(map[string]interface{})
	return r.RenderJSON(data)
}
