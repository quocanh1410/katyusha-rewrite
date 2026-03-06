// package junbanmachi // 順番待ち - じゅんばんまち - Queuing

// import "github.com/bonavadeur/katyusha/pkg/bonalib"

// func (q *ExtraQueue) SortAlgorithm(p *Packet) {
// 	bonalib.Info("SortAlgorithm", "Packet", p)
// 	// example of adding header
// 	p.Headers = append(p.Headers, &PushRequest_HeaderSchema{
// 		Field: "Katyusha-J-Field-1",
// 		Value: "Katyusha-J-Value-1",
// 	})

// 	q.Queue = append([]*Packet{p}, q.Queue...)
// }





// rewrite with latency-based sorting
package junbanmachi // 順番待ち - じゅんばんまち - Queuing

import (
	"time"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
	"github.com/bonavadeur/katyusha/pkg/common"
)

func (q *ExtraQueue) SortAlgorithm(p *Packet) {

	bonalib.Info("SortAlgorithm", "Packet", p)

	// DÙNG SourceIP LÀM KEY (hoặc X-Request-Id nếu có)
	key := p.SourceIP

	// LƯU START TIME GLOBAL
	if key != "" {
		common.GlobalRequestStart.Store(key, time.Now())
	}

	// example of adding header
	p.Headers = append(p.Headers, &PushRequest_HeaderSchema{
		Field: "Katyusha-J-Field-1",
		Value: "Katyusha-J-Value-1",
	})

	q.Queue = append([]*Packet{p}, q.Queue...)
}



