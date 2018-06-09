package quasar

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	pb "github.com/Bit-Nation/go-libp2p-quasar/pb"
	net "github.com/libp2p/go-libp2p-net"
	protoEnc "github.com/multiformats/go-multicodec/protobuf"
	uuid "github.com/satori/go.uuid"
)

type response struct {
	msg *pb.Message
	err error
}

type request struct {
	msg      *pb.Message
	respChan chan response
}

func (p *protocol) addRequest(req request) {

	p.lock.Lock()
	p.requests[req.msg.RequestID] = req
	p.lock.Unlock()

}

// cut request from internal stack
func (p *protocol) cutRequest(id string) (request, error) {

	p.lock.Lock()
	req, exist := p.requests[id]
	// in case it exist we want to delete it to free the map
	if exist {
		delete(p.requests, id)
	}
	p.lock.Unlock()

	if !exist {
		return request{}, errors.New(fmt.Sprintf("couldn't find request for ID: %s", id))
	}

	return req, nil

}

// respond to request
func (p *protocol) respond(id string, response *pb.Message, str net.Stream) error {

	w := bufio.NewWriter(str)
	enc := protoEnc.Multicodec(nil).Encoder(w)

	response.RequestID = id

	// encode msg and send it
	err := enc.Encode(response)
	if err != nil {
		return err
	}
	return w.Flush()

}

// send a message
func (p *protocol) request(msg *pb.Message, str net.Stream, timeOut time.Duration) (response, error) {

	w := bufio.NewWriter(str)
	enc := protoEnc.Multicodec(nil).Encoder(w)

	// resp chan
	c := make(chan response)

	// create request id
	id, err := uuid.NewV4()
	if err != nil {
		return response{}, err
	}

	msg.RequestID = id.String()

	// encode msg and send it
	err = enc.Encode(msg)
	if err != nil {
		return response{}, err
	}

	if err := w.Flush(); err != nil {
		return response{}, err
	}

	p.addRequest(request{
		msg:      msg,
		respChan: c,
	})

	select {
	case res := <-c:
		return res, nil
	case <-time.After(timeOut):
		// but the request from stack which will remove it
		p.cutRequest(msg.RequestID)
		return response{}, errors.New(fmt.Sprintf("request timeout for ID: %s", msg.RequestID))
	}

}
