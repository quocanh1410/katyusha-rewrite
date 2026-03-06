package hashi

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
    "github.com/bonavadeur/katyusha/pkg/bonalib"
	"google.golang.org/protobuf/proto"
)

type HalfAsyncHashi struct {
	Name           string
	bridgeType     BridgeType // ["client", "server"]
	upstreamFile   string
	downstreamFile string
	upstream       *os.File // Write
	downstream     *os.File // Read
	buffer         []byte
	requestSchema  reflect.Type
	responseSchema reflect.Type
	serverCallback BridgeCallback // nil for Client
	sendLock       sync.Mutex
	Bucket         chan proto.Message
}

func NewHalfAsyncHashi(
	name string,
	bridgeType BridgeType,
	upstreamFile string,
	downstreamFile string,
	requestSchema reflect.Type,
	responseSchema reflect.Type,
	serverCallback BridgeCallback,
) *HalfAsyncHashi {
	newHalfAsyncHashi := &HalfAsyncHashi{
		Name:           name,
		bridgeType:     bridgeType,
		upstreamFile:   upstreamFile,
		downstreamFile: downstreamFile,
		buffer:         make([]byte, 1024),
		requestSchema:  requestSchema,
		responseSchema: responseSchema,
		serverCallback: serverCallback,
		sendLock:       sync.Mutex{},
		Bucket:         make(chan proto.Message, 1),
	}

	var err error
	checkPipeExist(downstreamFile)
	checkPipeExist(upstreamFile)

	if bridgeType == HASHI_TYPE_HALF_ASYNC_SERVER {
		newHalfAsyncHashi.downstream, err = os.OpenFile(downstreamFile, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			panic(err)
		}
		newHalfAsyncHashi.upstream, err = os.OpenFile(upstreamFile, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			panic(err)
		}
		go newHalfAsyncHashi.AsyncReceiveServer()
	}
	if bridgeType == HASHI_TYPE_HALF_ASYNC_CLIENT {
		newHalfAsyncHashi.upstream, err = os.OpenFile(upstreamFile, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			panic(err)
		}
		newHalfAsyncHashi.downstream, err = os.OpenFile(downstreamFile, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			panic(err)
		}
		go newHalfAsyncHashi.AsyncReceiveClient()
	}

	return newHalfAsyncHashi
}

// func (hah *HalfAsyncHashi) AsyncSendClient(message proto.Message) (proto.Message, error) { // for Client
// 	// marshal message
// 	sentMessageBytes, err := proto.Marshal(message)
// 	if err != nil {
// 		log.Fatalln("Failed to encode sentMessage:", err)
// 		return nil, err
// 	}

// 	// send
// 	hah.sendLock.Lock()
// 	_, err = hah.upstream.Write(sentMessageBytes)
// 	if err != nil {
// 		fmt.Println("Error writing to upstream:", err)
// 		return nil, err
// 	}

// 	// receive
// 	receivedMessage := <-hah.Bucket
// 	hah.sendLock.Unlock()

// 	return receivedMessage, nil
// }


func (hah *HalfAsyncHashi) AsyncSendClient(message proto.Message) (proto.Message, error) {
    bonalib.Info("AsyncSendClient CALLED", "Worker", hah.Name)
	start := time.Now() // ⬅️ BẮT ĐẦU ĐO

	// marshal message
	sentMessageBytes, err := proto.Marshal(message)
	if err != nil {
		log.Fatalln("Failed to encode sentMessage:", err)
		return nil, err
	}

	// send
	hah.sendLock.Lock()
	_, err = hah.upstream.Write(sentMessageBytes)
	if err != nil {
		hah.sendLock.Unlock()
		fmt.Println("Error writing to upstream:", err)
		return nil, err
	}

	// receive (BLOCK Ở ĐÂY)
	receivedMessage := <-hah.Bucket
	hah.sendLock.Unlock()

	latency := time.Since(start) // ⬅️ KẾT THÚC ĐO

	bonalib.Info(
    "HashiLatency",
    "Worker", hah.Name,
    "UnixNano", start.UnixNano(),
    "LatencyMs", latency.Milliseconds(),
)

	return receivedMessage, nil
}

//gốc
func (hah *HalfAsyncHashi) AsyncReceiveClient() { // for Client
	// receive
	for {
		n, err := hah.downstream.Read(hah.buffer)
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalf("Error reading from FIFO: %v", err)
			panic(err)
		}

		receivedMessageBytes := hah.buffer[:n]

		receivedMessage := reflect.New(hah.responseSchema).Interface().(proto.Message)
		err = proto.Unmarshal(receivedMessageBytes, receivedMessage)
		if err != nil {
			log.Fatalf("AsyncReceiveClient Failed to unmarshal message: %v", err)
			panic(err)
		}

		hah.Bucket <- receivedMessage
	}
}

func (hah *HalfAsyncHashi) AsyncReceiveServer() error { // for Server
	// receive
	for {
		n, err := hah.downstream.Read(hah.buffer)
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalf("Error reading from FIFO: %v", err)
			return err
		}

		receivedMessageBytes := hah.buffer[:n]

		receivedMessage := reflect.New(hah.requestSchema).Interface().(proto.Message)
		err = proto.Unmarshal(receivedMessageBytes, receivedMessage)
		if err != nil {
			log.Fatalf("AsyncReceiveServer Failed to unmarshal message: %v", err)
			return err
		}

		hah.AsyncSendServer(receivedMessage)
	}
}

func (hah *HalfAsyncHashi) AsyncSendServer(message proto.Message) { // for Server
	// run callback function
	result, _ := hah.serverCallback(message)

	// marshal message
	responseMessageBytes, err := proto.Marshal(result.(proto.Message))
	if err != nil {
		log.Fatalln("Failed to encode sentMessage:", err)
		panic(err)
	}

	// send
	hah.sendLock.Lock()
	_, err = hah.upstream.Write(responseMessageBytes)
	if err != nil {
		fmt.Println("Error writing to upstream:", err)
		panic(err)
	}
	hah.sendLock.Unlock()
}
