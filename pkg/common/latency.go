// package common

// import( "sync"
//          "time"
// )

// var GlobalRequestStart sync.Map 
// var GlobalRequestTarget sync.Map
// var PodInFlight sync.Map

// type PodMetrics struct {
// 	Count int
// 	Total time.Duration
// }

// var PodMetric sync.Map

package common

import (
	"sync"
	"time"
)

// ================= GLOBAL =================

var GlobalRequestStart sync.Map
var GlobalRequestTarget sync.Map
var PodInFlight sync.Map

// ================= METRIC =================

type PodMetrics struct {
	Count int
	Total time.Duration
}

var PodMetric sync.Map

// ================= COUNTER =================

// 🔥 Export field để package khác dùng được
type Counter struct {
	Mu  sync.Mutex
	Val int
}

// 🔥 Dùng chung cho mọi package
func GetCounter(target string) *Counter {
	v, _ := PodInFlight.LoadOrStore(target, &Counter{})
	return v.(*Counter)
}