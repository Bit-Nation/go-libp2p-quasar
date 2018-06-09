package quasar

import (
	"errors"
	"fmt"
	"time"
	
	atbf "github.com/Bit-Nation/go-libp2p-quasar/atbf"
	pb "github.com/Bit-Nation/go-libp2p-quasar/pb"
	net "github.com/libp2p/go-libp2p-net"
)

// fetch filters for given stream
func (p *protocol) pullFilters(str net.Stream, timeOut time.Duration) (*pb.PullFiltersResponse, error) {

	m := pb.Message{
		Type: pb.Message_PULL_FILTER_REQUEST,
	}

	resp, err := p.request(&m, str, timeOut)
	if err != nil {
		return nil, err
	}

	if resp.msg.Type != pb.Message_PULL_FILTER_RESPONSE {
		return nil, errors.New(fmt.Sprintf("expected to get a PULL_FILTER_RESPONSE but got: %d", resp.msg.Type))
	}

	return resp.msg.PullFiltersResponse, nil

}

// push filter to a given stream / node
func (p *protocol) pushFilters(str net.Stream, timeOut time.Duration, filter *atbf.AttenuatedBloomFilter) error {

	rawFilter, err := filter.Marshal()
	if err != nil {
		return err
	}

	m := pb.Message{
		Type: pb.Message_PUSH_FILTER_REQUEST,
		PushFilterRequest: &pb.PushFiltersRequest{
			AttenuateBloomFilter: rawFilter,
		},
	}

	resp, err := p.request(&m, str, timeOut)
	if err != nil {
		return err
	}

	if resp.msg.Type != pb.Message_PUSH_FILTER_RESPONE {
		return errors.New(fmt.Sprintf("expected to get a PUSH_FILTER_RESPONE but got: %d", resp.msg.Type))
	}

	return nil

}

// handle filter request
func (p *protocol) handlePullFilterRequest(msg *pb.Message, str net.Stream) error {

	// check if message is correct
	if msg.Type != pb.Message_PULL_FILTER_REQUEST {
		return errors.New(fmt.Sprintf("got %s instead of PULL_FILTER_REQUEST", msg.Type))
	}

	// marshal the whole thing
	rawFilter, err := p.filter.Marshal()
	if err != nil {
		return err
	}

	// send response to back to peer
	return p.respond(msg.RequestID, &pb.Message{
		Type: pb.Message_PULL_FILTER_RESPONSE,
		PullFiltersResponse: &pb.PullFiltersResponse{
			AttenuateBloomFilter: rawFilter,
		},
	}, str)

}

// handle pull filter response
func (p *protocol) handlePullFilterResponse(msg *pb.Message) error {

	// check if message is correct
	if msg.Type != pb.Message_PULL_FILTER_RESPONSE {
		return errors.New(fmt.Sprintf("got %s instead of PULL_FILTER_RESPONSE", msg.Type))
	}

	req, err := p.cutRequest(msg.RequestID)
	if err != nil {
		return err
	}

	req.respChan <- response{
		msg: msg,
	}

	return nil
}

// handle push filter response
func (p *protocol) handlePushFilterResponse(msg *pb.Message) error {

	// check if message is correct
	if msg.Type != pb.Message_PUSH_FILTER_RESPONE {
		return errors.New(fmt.Sprintf("got %s instead of PUSH_FILTER_RESPONE", msg.Type))
	}

	req, err := p.cutRequest(msg.RequestID)
	if err != nil {
		return err
	}

	req.respChan <- response{
		msg: msg,
	}

	return nil

}

// handle push filter request
func (p *protocol) handlePushFilterRequest(msg *pb.Message, str net.Stream) error {

	// check if message is correct
	if msg.Type != pb.Message_PUSH_FILTER_REQUEST {
		return errors.New(fmt.Sprintf("got %s instead of PUSH_FILTER_REQUEST", msg.Type))
	}

	// merge our filter with the received one
	recFilter := &atbf.AttenuatedBloomFilter{}
	if err := recFilter.Unmarshal(msg.PushFilterRequest.AttenuateBloomFilter); err != nil {
		return err
	}
	if err := p.filter.Merge(recFilter); err != nil {
		return err
	}

	// send response to back to peer
	return p.respond(msg.RequestID, &pb.Message{
		Type:               pb.Message_PUSH_FILTER_RESPONE,
		PushFilterResponse: &pb.PushFiltersResponse{},
	}, str)

}
