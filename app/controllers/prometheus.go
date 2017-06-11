package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	AppPageHitIndex = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_index",
		Help: "Number of page hits",
	})

	AppPageHitFAQ = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_faq",
		Help: "Number of page hits",
	})

	AppPageHitDonate = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_donate",
		Help: "Number of page hits",
	})

	AppPageHitInformation = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_information",
		Help: "Number of page hits",
	})

	AppPageHitContact = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_contact",
		Help: "Number of page hits",
	})

	AppPageHitLanding = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_landing",
		Help: "Number of page hits",
	})

	AppPageHitLogin = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_login",
		Help: "Number of page hits",
	})

	AppPageHitRegister = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_register",
		Help: "Number of page hits",
	})

	AppPageHitVerifyEmail = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_verifyemail",
		Help: "Number of page hits",
	})

	AppPageHitNewPassGet = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_newpassget",
		Help: "Number of page hits",
	})

	AppPageHitNewPassPost = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_newpasspost",
		Help: "Number of page hits",
	})

	AppPageHitDashboard = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_dashboard",
		Help: "Number of page hits",
	})

	AppPageHitInfoDashboard = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_infodash",
		Help: "Number of page hits",
	})

	AppPageHitAdvInfoDashboard = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_advinfodash",
		Help: "Number of page hits",
	})

	AppPageHitInfoLogout = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_logout",
		Help: "Number of page hits",
	})

	AppPageHitInfoEnable2fa = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_2faenable",
		Help: "Number of page hits",
	})

	AppPageHitSetPoloKeys = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_setpolokeys",
		Help: "Number of page hits",
	})

	AppPageHitSetSettingDashUser = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_usersetting",
		Help: "Number of page hits",
	})

	AppPageHitSetSettingDashLend = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_userlend",
		Help: "Number of page hits",
	})

	AppPageHitCreate2fa = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_2facreate",
		Help: "Number of page hits",
	})

	AppPageHitEnableUserLending = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_enablelending",
		Help: "Number of page hits",
	})

	AppPageHitSetEnableUserLending = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_setenablelending",
		Help: "Number of page hits",
	})

	AppPageHitReqEmailVerify = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_reqverifyemail",
		Help: "Number of page hits",
	})

	AppPageHitChangePass = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_changepass",
		Help: "Number of page hits",
	})

	AppPageAuthUser = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_controller_page_hit_authenticate",
		Help: "Number of page hits",
	})
)

var registered bool = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true

	prometheus.MustRegister(AppPageHitIndex)
	prometheus.MustRegister(AppPageHitFAQ)
	prometheus.MustRegister(AppPageHitDonate)
	prometheus.MustRegister(AppPageHitInformation)
	prometheus.MustRegister(AppPageHitContact)
	prometheus.MustRegister(AppPageHitLanding)
	prometheus.MustRegister(AppPageHitLogin)
	prometheus.MustRegister(AppPageHitRegister)
	prometheus.MustRegister(AppPageHitVerifyEmail)
	prometheus.MustRegister(AppPageHitNewPassGet)
	prometheus.MustRegister(AppPageHitNewPassPost)
	prometheus.MustRegister(AppPageHitDashboard)
	prometheus.MustRegister(AppPageHitInfoDashboard)
	prometheus.MustRegister(AppPageHitAdvInfoDashboard)
	prometheus.MustRegister(AppPageHitInfoLogout)
	prometheus.MustRegister(AppPageHitInfoEnable2fa)
	prometheus.MustRegister(AppPageHitSetPoloKeys)
	prometheus.MustRegister(AppPageHitSetSettingDashUser)
	prometheus.MustRegister(AppPageHitSetSettingDashLend)
	prometheus.MustRegister(AppPageHitCreate2fa)
	prometheus.MustRegister(AppPageHitEnableUserLending)
	prometheus.MustRegister(AppPageHitSetEnableUserLending)
	prometheus.MustRegister(AppPageHitReqEmailVerify)
	prometheus.MustRegister(AppPageHitChangePass)
	prometheus.MustRegister(AppPageAuthUser)
}
