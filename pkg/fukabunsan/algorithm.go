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



// // rewrite with Power of Two Choices and latency-based selection
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
// package fukabunsan

// import (
// 	"math/rand"
// 	"time"

// 	"github.com/bonavadeur/katyusha/pkg/bonalib"
// 	"github.com/bonavadeur/katyusha/pkg/common"
// )

// // ================= LATENCY =================

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

// // ================= LB =================

// func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {

// 	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)

// 	if len(lbRequest.Targets) == 0 {
// 		return nil
// 	}

// 	var selectedTarget string

// 	// ================= CHỌN POD =================

// 	if len(lbRequest.Targets) == 1 {

// 		selectedTarget = lbRequest.Targets[0]

// 		// 🔥 atomic check + set
// 		counter := common.GetCounter(selectedTarget)
// 		counter.Mu.Lock()
// 		if counter.Val != 0 {
// 			counter.Mu.Unlock()
// 			return nil
// 		}
// 		counter.Val++
// 		counter.Mu.Unlock()

// 	} else {

// 		// Power of Two Choices
// 		i := rand.Intn(len(lbRequest.Targets))
// 		j := rand.Intn(len(lbRequest.Targets))

// 		for j == i {
// 			j = rand.Intn(len(lbRequest.Targets))
// 		}

// 		t1 := lbRequest.Targets[i]
// 		t2 := lbRequest.Targets[j]

// 		// 🔥 thử chọn t1 trước
// 		counter1 := common.GetCounter(t1)
// 		counter1.Mu.Lock()
// 		if counter1.Val == 0 {
// 			counter1.Val++
// 			selectedTarget = t1
// 			counter1.Mu.Unlock()
// 		} else {
// 			counter1.Mu.Unlock()

// 			// 🔥 thử t2
// 			counter2 := common.GetCounter(t2)
// 			counter2.Mu.Lock()
// 			if counter2.Val == 0 {
// 				counter2.Val++
// 				selectedTarget = t2
// 				counter2.Mu.Unlock()
// 			} else {
// 				counter2.Mu.Unlock()
// 				return nil // ❌ cả 2 đều bận
// 			}
// 		}

// 		// 🔥 nếu cả 2 đều rảnh → chọn theo latency
// 		if selectedTarget == "" {

// 			l1 := getAvgLatency(t1)
// 			l2 := getAvgLatency(t2)

// 			if l1 == 0 {
// 				selectedTarget = t1
// 			} else if l2 == 0 {
// 				selectedTarget = t2
// 			} else if l1 <= l2 {
// 				selectedTarget = t1
// 			} else {
// 				selectedTarget = t2
// 			}

// 			counter := common.GetCounter(selectedTarget)
// 			counter.Mu.Lock()
// 			counter.Val++
// 			counter.Mu.Unlock()
// 		}
// 	}

// 	if selectedTarget == "" {
// 		bonalib.Info("❌ NO AVAILABLE POD")
// 		return nil
// 	}

// 	bonalib.Info("🚀 ASSIGN:", selectedTarget)

// 	// ================= LƯU MAPPING =================

// 	key := lbRequest.SourceIP
// 	if key != "" {
// 		common.GlobalRequestTarget.Store(key, selectedTarget)
// 	}

// 	// ================= RESPONSE =================

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
/////////////////////////////////////////////////////////////
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

// ================= LB =================
func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {

	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)

	if len(lbRequest.Targets) == 0 {
		return nil
	}

	// ================= TÌM POD RẢNH =================
	var candidates []string
	for _, target := range lbRequest.Targets {
		counter := common.GetCounter(target)
		counter.Mu.Lock()
		if counter.Val == 0 {
			candidates = append(candidates, target)
		}
		counter.Mu.Unlock()
	}

	// ================= XỬ LÝ =================
	if len(candidates) == 0 {
		// tất cả pod bận → drop request
		bonalib.Info("❌ ALL PODS BUSY")
		return nil
	}

	// chọn pod rảnh có latency thấp nhất
	selectedTarget := candidates[0]
	minLatency := getAvgLatency(selectedTarget)

	for _, target := range candidates[1:] {
		l := getAvgLatency(target)
		if l < minLatency {
			minLatency = l
			selectedTarget = target
		}
	}

	// tăng counter pod được chọn
	counter := common.GetCounter(selectedTarget)
	counter.Mu.Lock()
	counter.Val++
	counter.Mu.Unlock()

	bonalib.Info("🚀 ASSIGN:", selectedTarget)

	// ================= LƯU MAPPING =================
	key := lbRequest.SourceIP
	if key != "" {
		common.GlobalRequestTarget.Store(key, selectedTarget)
	}

	// ================= RESPONSE =================
	ret := &LBResponse{
		Target:  selectedTarget,
		Headers: []*LBResponse_HeaderSchema{{Field: "Katyusha-F-Field", Value: "Katyusha-F-Field"}},
	}

	return ret
}