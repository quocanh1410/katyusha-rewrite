// package fukabunsan // 負荷分散 - ふかぶんさん - Load Balancing

// import (
// 	"github.com/bonavadeur/katyusha/pkg/bonalib"
	
// )

// func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {
// 	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)
	
// 	// return lbRequest.Targets[0]
// 	ret := &LBResponse{
// 		Target:  lbRequest.Targets[0],
// 		Headers: make([]*LBResponse_HeaderSchema, 0),
// 	}
// 	ret.Headers = append(ret.Headers, &LBResponse_HeaderSchema{
// 		Field: "Katyusha-F-Field",
// 		Value: "Katyusha-F-Field",
// 	})

// 	return ret
// }



// rewrite with Power of Two Choices and latency-based selection
package fukabunsan

import (
	"math/rand"
	"time"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
	"github.com/bonavadeur/katyusha/pkg/common"
)

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

func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {

	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)

	if len(lbRequest.Targets) == 0 {
		return nil
	}

	var selectedTarget string

	// Nếu chỉ có 1 pod
	if len(lbRequest.Targets) == 1 {
		selectedTarget = lbRequest.Targets[0]

	} else {

		// Power of Two Choices
		i := rand.Intn(len(lbRequest.Targets))
		j := rand.Intn(len(lbRequest.Targets))

		for j == i {
			j = rand.Intn(len(lbRequest.Targets))
		}

		t1 := lbRequest.Targets[i]
		t2 := lbRequest.Targets[j]

		l1 := getAvgLatency(t1)
		l2 := getAvgLatency(t2)

		if l1 == 0 {
			selectedTarget = t1
		} else if l2 == 0 {
			selectedTarget = t2
		} else if l1 <= l2 {
			selectedTarget = t1
		} else {
			selectedTarget = t2
		}
	}

	// Lưu SourceIP → Pod
	key := lbRequest.SourceIP
	if key != "" {
		common.GlobalRequestTarget.Store(key, selectedTarget)
		bonalib.Info("🔥 SAVE TARGET:", key, "→", selectedTarget)
	}

	// Trả response
	ret := &LBResponse{
		Target:  selectedTarget,
		Headers: make([]*LBResponse_HeaderSchema, 0),
	}

	ret.Headers = append(ret.Headers, &LBResponse_HeaderSchema{
		Field: "Katyusha-F-Field",
		Value: "Katyusha-F-Field",
	})

	return ret
}