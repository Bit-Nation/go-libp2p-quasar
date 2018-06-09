package quasar

import (
	"bufio"

	pb "github.com/Bit-Nation/go-libp2p-quasar/pb"
	net "github.com/libp2p/go-libp2p-net"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
)

// handle my stream
func (p *protocol) streamHandler(str net.Stream) {
	go func(p *protocol) {
		r := bufio.NewReader(str)
		dec := protobufCodec.Multicodec(nil).Decoder(r)
		// @todo I think as peers come and go this loop needs to be "broken". I am not sure if lp2p handles it.
		for {
			var msg pb.Message
			err := dec.Decode(&msg)
			if err != nil {
				logger.Error(err)
			}
			logger.Info("Got message: ", msg.RequestID, " ", msg.Type)
			switch os := msg.Type; os {
			case pb.Message_PULL_FILTER_RESPONSE:
				if err := p.handlePullFilterResponse(&msg); err != nil {
					logger.Error(err)
				}
			case pb.Message_PULL_FILTER_REQUEST:
				if err := p.handlePullFilterRequest(&msg, str); err != nil {
					logger.Error(err)
				}
			case pb.Message_PUSH_FILTER_RESPONE:
				if err := p.handlePushFilterResponse(&msg); err != nil {
					logger.Error(err)
				}
			case pb.Message_PUSH_FILTER_REQUEST:
				if err := p.handlePushFilterRequest(&msg, str); err != nil {
					logger.Error(err)
				}
			default:
				logger.Error("couldn't handle: ", os)
			}

		}
	}(p)
}
