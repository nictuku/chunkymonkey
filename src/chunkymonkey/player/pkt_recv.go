package player

import (
	"io"
	"log"
	"os"

	"chunkymonkey/proto"
)

// playerRx deals with receiving packets from the client.
type playerRx struct {
	pktSerial proto.PacketSerializer
	conn      io.Reader

	ctrl chan struct{}

	recvPkt chan<- interface{}
	RecvPkt <-chan interface{}

	recvErr chan<- os.Error
	RecvErr <-chan os.Error
}

func (p *playerRx) init(conn io.Reader) {
	p.conn = conn

	p.ctrl = make(chan struct{})

	recvPkt := make(chan interface{})
	p.recvPkt = recvPkt
	p.RecvPkt = recvPkt

	// Error channel can hold one so that we can exit the goroutine without
	// blocking.
	recvErr := make(chan os.Error, 1)
	p.recvErr = recvErr
	p.RecvErr = recvErr
}

func (p *playerRx) Stop() {
	close(p.ctrl)
}

func (p *playerRx) loop() {
	for {
		if pkt, err := p.pktSerial.ReadPacket(p.conn, true); err != nil {
			p.recvErr <- err
			return
		} else {
			select {
			case p.recvPkt <- pkt:
			case _, ok := <-p.ctrl:
				if !ok {
					return
				} else {
					log.Print("warning - unhandled control message to playerRx")
				}
			}
		}
	}
}
