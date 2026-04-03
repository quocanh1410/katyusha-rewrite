package fukabunsan

type QueuedRequest struct {
	req    *LBRequest
	respCh chan *LBResponse
}