package outoushuugou // 応答集合 - おうとうしゅうごう - Response Pool

import (
	reflect "reflect"
	"sync"
	"time"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
	"github.com/bonavadeur/katyusha/pkg/hashi"
	"github.com/bonavadeur/katyusha/pkg/common"
	
)

type ResponsePool struct {
	responseBridge    *hashi.Hashi
	Pool              []*ResponseFeedback
	PoolAppendingLock *sync.Mutex
	RequestID         string
}

func NewResponsePool() *ResponsePool {
	newResponsePool := &ResponsePool{
		Pool:              make([]*ResponseFeedback, 0),
		PoolAppendingLock: &sync.Mutex{},
	}

	newResponsePool.responseBridge = hashi.NewHashi(
		"responsePoolBridge",
		hashi.HASHI_TYPE_SERVER,
		BASE_PATH+"/response-pool-bridge",
		bonalib.Cm2Int("katyusha-threads"),
		reflect.TypeOf(ResponseFeedback{}),
		reflect.TypeOf(ResponseConfirm{}),
		newResponsePool.ResponsePoolAdapter,
	)

	return newResponsePool
}

// func (rp *ResponsePool) ResponsePoolAdapter(params ...interface{}) (interface{}, error) {
// 	feedback := params[0].(*ResponseFeedback)

// 	rp.PoolAppendingLock.Lock()
// 	bonalib.Info("ResponsePoolAdapter", feedback)
// 	rp.Pool = append([]*ResponseFeedback{feedback}, rp.Pool...)
// 	rp.PoolAppendingLock.Unlock()

// 	return &ResponseConfirm{SymbolizeResponse: Status_Success}, nil
// }




//rewrite with latency-based metric collection
func (rp *ResponsePool) ResponsePoolAdapter(params ...interface{}) (interface{}, error) {

	feedback := params[0].(*ResponseFeedback)

	// DÙNG SourceIP LÀM KEY
	key := feedback.SourceIP

	if key != "" {

		// LẤY POD TARGET ĐÃ LƯU Ở LBAlgorithm
		var target string
		if value, ok := common.GlobalRequestTarget.Load(key); ok {
			target = value.(string)
			common.GlobalRequestTarget.Delete(key)
		}

		// LẤY START TIME VÀ TÍNH LATENCY
		if value, ok := common.GlobalRequestStart.Load(key); ok {

			startTime := value.(time.Time)
			latency := time.Since(startTime)

			bonalib.Info("🔥 POD:", target, "LATENCY:", latency)

			// CẬP NHẬT METRIC CỦA POD
			if target != "" {

				metricValue, _ := common.PodMetric.LoadOrStore(target, &common.PodMetrics{})
				metric := metricValue.(*common.PodMetrics)

				metric.Count++
				metric.Total += latency
			}

			common.GlobalRequestStart.Delete(key)
		}
	}

	// Giữ nguyên phần pool
	rp.PoolAppendingLock.Lock()
	bonalib.Info("ResponsePoolAdapter", feedback)
	rp.Pool = append([]*ResponseFeedback{feedback}, rp.Pool...)
	rp.PoolAppendingLock.Unlock()

	return &ResponseConfirm{SymbolizeResponse: Status_Success}, nil
}