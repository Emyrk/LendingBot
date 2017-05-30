// GENERATED CODE - DO NOT EDIT
package routes

import "github.com/revel/revel"


type tApp struct {}
var App tApp


func (_ tApp) Sandbox(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("App.Sandbox", args).URL
}

func (_ tApp) Index(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("App.Index", args).URL
}

func (_ tApp) Landing(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("App.Landing", args).URL
}

func (_ tApp) Login(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("App.Login", args).URL
}

func (_ tApp) Register(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("App.Register", args).URL
}


type tAppAuthRequired struct {}
var AppAuthRequired tAppAuthRequired


func (_ tAppAuthRequired) Dashboard(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.Dashboard", args).URL
}

func (_ tAppAuthRequired) Logout(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.Logout", args).URL
}

func (_ tAppAuthRequired) InfoDashboard(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.InfoDashboard", args).URL
}

func (_ tAppAuthRequired) Enable2FA(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.Enable2FA", args).URL
}

func (_ tAppAuthRequired) InfoAdvancedDashboard(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.InfoAdvancedDashboard", args).URL
}

func (_ tAppAuthRequired) SettingsDashboard(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.SettingsDashboard", args).URL
}

func (_ tAppAuthRequired) Create2FA(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.Create2FA", args).URL
}

func (_ tAppAuthRequired) AuthUser(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppAuthRequired.AuthUser", args).URL
}


type tAppSysAdmin struct {}
var AppSysAdmin tAppSysAdmin


func (_ tAppSysAdmin) LogsDashboard(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppSysAdmin.LogsDashboard", args).URL
}

func (_ tAppSysAdmin) ExportLogs(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppSysAdmin.ExportLogs", args).URL
}

func (_ tAppSysAdmin) DeleteLogs(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppSysAdmin.DeleteLogs", args).URL
}

func (_ tAppSysAdmin) AuthUserSysAdmin(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("AppSysAdmin.AuthUserSysAdmin", args).URL
}


type tStatic struct {}
var Static tStatic


func (_ tStatic) Serve(
		prefix string,
		filepath string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "prefix", prefix)
	revel.Unbind(args, "filepath", filepath)
	return revel.MainRouter.Reverse("Static.Serve", args).URL
}

func (_ tStatic) ServeModule(
		moduleName string,
		prefix string,
		filepath string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "moduleName", moduleName)
	revel.Unbind(args, "prefix", prefix)
	revel.Unbind(args, "filepath", filepath)
	return revel.MainRouter.Reverse("Static.ServeModule", args).URL
}


type tTestRunner struct {}
var TestRunner tTestRunner


func (_ tTestRunner) Index(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("TestRunner.Index", args).URL
}

func (_ tTestRunner) Suite(
		suite string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "suite", suite)
	return revel.MainRouter.Reverse("TestRunner.Suite", args).URL
}

func (_ tTestRunner) Run(
		suite string,
		test string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "suite", suite)
	revel.Unbind(args, "test", test)
	return revel.MainRouter.Reverse("TestRunner.Run", args).URL
}

func (_ tTestRunner) List(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("TestRunner.List", args).URL
}


