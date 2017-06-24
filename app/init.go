package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Emyrk/LendingBot/app/controllers"
	// "github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/revel/revel"
	log "github.com/sirupsen/logrus"
)

// var _ = mongo.CreateMongoDB

var cLog = log.WithField("package", "init")

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.ActionInvoker,           // Invoke the action.
	}
	revel.InterceptMethod(controllers.App.AppAuthUser, revel.BEFORE)
	revel.InterceptMethod(controllers.AppAuthRequired.AuthUser, revel.BEFORE)
	revel.InterceptMethod(controllers.AppSysAdmin.AuthUserSysAdmin, revel.BEFORE)
	revel.InterceptMethod(controllers.AppAdmin.AuthUserAdmin, revel.BEFORE)
	// register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)

	// Revel custom funcs
	revel.TemplateFuncs["floateq"] = func(a float64, b float64) bool {
		return a == b
	}

	revel.TemplateFuncs["floatge"] = func(a float64, b float64) bool {
		return a > b
	}

	revel.TemplateFuncs["isEven"] = func(a int) bool {
		return a%2 == 0
	}

	revel.TemplateFuncs["contains"] = func(s string, substr string) bool {
		return strings.Contains(s, substr)
	}

	revel.TemplateFuncs["formatPercentString"] = func(a string, precision int) string {
		s, err := strconv.ParseFloat(a, 64)
		if err != nil {
			fmt.Println(err)
			return a
		}
		s = s * 100
		return formatFloat(s, precision)
	}

	revel.TemplateFuncs["formatFloat"] = func(a float64, precision int) string {
		return formatFloat(a, precision)
	}

	revel.TemplateFuncs["formatFloatPercent"] = func(a float64, precision int) string {
		a = a * 100
		return formatFloat(a, precision)
	}

	revel.OnAppStart(controllers.Launch)

	// revel. .OnAppShutdown(controllers.Shutdown)
}

func formatFloat(a float64, precision int) string {
	switch precision {
	case 1:
		return fmt.Sprintf("%.1f", a)
	case 2:
		return fmt.Sprintf("%.2f", a)
	case 3:
		return fmt.Sprintf("%.3f", a)
	case 4:
		return fmt.Sprintf("%.4f", a)
	case 5:
		return fmt.Sprintf("%.5f", a)
	}
	return fmt.Sprintf("%f", a)
}

// HeaderFilter adds common security headers
// TODO turn this into revel.HeaderFilter
// should probably also have a filter for CSRF
// not sure if it can go in the same filter or not
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}

func InitDB() {
	// llog := cLog.WithField("method", "InitDB")
	// // The second argument are default values, for safety
	// uri := revel.Config.StringDefault("database.uri", "mongodb://localhost:27017")
	// name := revel.Config.StringDefault("database.name", "LendingBot")
	// if err := mongo.Init(uri, name); err != nil {
	// 	llog.Errorf("InitMongo: %s\n", err.Error())
	// }
}
