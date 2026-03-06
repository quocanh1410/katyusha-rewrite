package common

import( "sync"
         "time"
)

var GlobalRequestStart sync.Map 
var GlobalRequestTarget sync.Map

type PodMetrics struct {
	Count int
	Total time.Duration
}

var PodMetric sync.Map




