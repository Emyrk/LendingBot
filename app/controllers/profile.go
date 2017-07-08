package controllers

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	// "runtime"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler() {
	// runtime.MemProfileRate = mpr
	log.Println(http.ListenAndServe("localhost:6066", nil))
	//runtime.SetBlockProfileRate(100000)
}
