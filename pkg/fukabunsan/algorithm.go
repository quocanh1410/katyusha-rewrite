// // package fukabunsan // 負荷分散 - ふかぶんさん - Load Balancing

// // import (
// // 	"github.com/bonavadeur/katyusha/pkg/bonalib"
	
// // )

// // func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {
// // 	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)
	
// // 	// return lbRequest.Targets[0]
// // 	ret := &LBResponse{
// // 		Target:  lbRequest.Targets[0],
// // 		Headers: make([]*LBResponse_HeaderSchema, 0),
// // 	}
// // 	ret.Headers = append(ret.Headers, &LBResponse_HeaderSchema{
// // 		Field: "Katyusha-F-Field",
// // 		Value: "Katyusha-F-Field",
// // 	})

// // 	return ret
// // }



// version 1

// package fukabunsan

// import (
// 	"math/rand"
// 	"time"

// 	"github.com/bonavadeur/katyusha/pkg/bonalib"
// 	"github.com/bonavadeur/katyusha/pkg/common"
// )

// func getAvgLatency(target string) time.Duration {

// 	value, ok := common.PodMetric.Load(target)
// 	if !ok {
// 		return 0
// 	}

// 	metric := value.(*common.PodMetrics)

// 	if metric.Count == 0 {
// 		return 0
// 	}

// 	return metric.Total / time.Duration(metric.Count)
// }

// func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {

// 	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)

// 	if len(lbRequest.Targets) == 0 {
// 		return nil
// 	}

// 	var selectedTarget string

// 	// Nếu chỉ có 1 pod
// 	if len(lbRequest.Targets) == 1 {
// 		selectedTarget = lbRequest.Targets[0]

// 	} else {

// 		// Power of Two Choices
// 		i := rand.Intn(len(lbRequest.Targets))
// 		j := rand.Intn(len(lbRequest.Targets))

// 		for j == i {
// 			j = rand.Intn(len(lbRequest.Targets))
// 		}

// 		t1 := lbRequest.Targets[i]
// 		t2 := lbRequest.Targets[j]

// 		l1 := getAvgLatency(t1)
// 		l2 := getAvgLatency(t2)

// 		if l1 == 0 {
// 			selectedTarget = t1
// 		} else if l2 == 0 {
// 			selectedTarget = t2
// 		} else if l1 <= l2 {
// 			selectedTarget = t1
// 		} else {
// 			selectedTarget = t2
// 		}
// 	}

// 	// Lưu SourceIP → Pod
// 	key := lbRequest.SourceIP
// 	if key != "" {
// 		common.GlobalRequestTarget.Store(key, selectedTarget)
// 		bonalib.Info("🔥 SAVE TARGET:", key, "→", selectedTarget)
// 	}

// 	//  Trả response
// 	ret := &LBResponse{
// 		Target:  selectedTarget,
// 		Headers: make([]*LBResponse_HeaderSchema, 0),
// 	}

// 	ret.Headers = append(ret.Headers, &LBResponse_HeaderSchema{
// 		Field: "Katyusha-F-Field",
// 		Value: "Katyusha-F-Field",
// 	})

// 	return ret
// }
/////////////////////////////////////////////////////
//version2 - latency-based sorting with queue and worker

package fukabunsan

import (
	"time"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
	"github.com/bonavadeur/katyusha/pkg/common"
)


// ================= LATENCY =================
func getAvgLatency(target string) time.Duration {
	value, ok := common.PodMetric.Load(target)
	if !ok {
		return 0
	}

	metric := value.(*common.PodMetrics)
	if metric.Count == 0 {
		return 0
	}

	return metric.Total / time.Duration(metric.Count)
}

// ================= LB (Producer) =================
func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {

	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)

	if len(lbRequest.Targets) == 0 {
		return nil
	}

	// 🔥 tạo request đưa vào queue
	qr := &QueuedRequest{
		req:    lbRequest,
		respCh: make(chan *LBResponse, 1),
	}

	// 🔥 push vào queue
	select {
	case lb.queue <- qr:
	default:
		bonalib.Info("❌ QUEUE FULL - DROP REQUEST")	
		return nil
	}

	// 🔥 đợi worker xử lý
	return <-qr.respCh
}

// ================= WORKER (Consumer) =================
func (lb *LoadBalancer) worker() {

	for qr := range lb.queue {

		lbRequest := qr.req
		var selectedTarget string

		for {
			var candidates []string

			// 🔍 tìm pod rảnh
			for _, target := range lbRequest.Targets {
				counter := common.GetCounter(target)

				counter.Mu.Lock()
				if counter.Val == 0 {
					candidates = append(candidates, target)
				}
				counter.Mu.Unlock()
			}

			// ✅ nếu có pod rảnh
			if len(candidates) > 0 {

				// 🔥 chọn pod latency thấp nhất
				selectedTarget = candidates[0]
				minLatency := getAvgLatency(selectedTarget)

				for _, target := range candidates[1:] {
					l := getAvgLatency(target)
					if l < minLatency {
						minLatency = l
						selectedTarget = target
					}
				}

				// 🔒 tăng counter
				counter := common.GetCounter(selectedTarget)
				counter.Mu.Lock()
				counter.Val++
				counter.Mu.Unlock()

				break
			}

			// ⏳ chờ pod rảnh (không drop nữa)
			time.Sleep(5 * time.Millisecond)
		}

		bonalib.Info("🚀 ASSIGN:", selectedTarget)

		// ================= LƯU MAPPING =================
		key := lbRequest.SourceIP
		if key != "" {
			common.GlobalRequestTarget.Store(key, selectedTarget)
		}

		// ================= RESPONSE =================
		qr.respCh <- &LBResponse{
			Target: selectedTarget,
			Headers: []*LBResponse_HeaderSchema{
				{Field: "Katyusha-F-Field", Value: "Katyusha-F-Field"},
			},
		}
	}
}