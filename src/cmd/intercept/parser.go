package main

import (
	"encoding/hex"
	"io"
	"log"
	"os"

	"chunkymonkey/proto"
)

// Hex dumps the input to the log
func (p *MessageParser) dumpInput(logPrefix string, reader io.Reader) {
	buf := make([]byte, 16, 16)
	for {
		_, err := io.ReadAtLeast(reader, buf, 1)
		if err != nil {
			return
		}

		hexData := hex.EncodeToString(buf)
		p.printf("Unparsed data: %s", hexData)
	}
}

// Consumes data from reader until an error occurs
func (p *MessageParser) consumeUnrecognizedInput(reader io.Reader) {
	p.printf("Lost packet sync. Ignoring further data.")
	buf := make([]byte, 4096)
	for {
		_, err := io.ReadAtLeast(reader, buf, 1)
		if err != nil {
			return
		}
	}
}

type MessageParser struct {
	logger     *log.Logger
	fromClient bool
	ps         proto.PacketSerializer
}

func NewMessageParser(logger *log.Logger, fromClient bool) *MessageParser {
	return &MessageParser{
		logger:     logger,
		fromClient: fromClient,
	}
}

func (p *MessageParser) printf(format string, v ...interface{}) {
	p.logger.Printf(format, v...)
}

func (p *MessageParser) Parse(reader io.Reader) {
	// If we return, we should consume all input to avoid blocking the pipe
	// we're listening on. TODO Maybe we could just close it?
	defer p.consumeUnrecognizedInput(reader)

	defer func() {
		if err := recover(); err != nil {
			p.printf("Parsing failed: %v", err)
		}
	}()

	for {
		pkt, err := p.ps.ReadPacket(reader, p.fromClient)
		if err != nil {
			if err != os.EOF {
				p.printf("ReceiveLoop failed: %v", err)
			} else {
				p.printf("ReceiveLoop hit EOF")
			}
			return
		} else {
			switch pktTyped := pkt.(type) {
			case *proto.PacketMapChunk:
				p.printf(
					"%T{Corner: %#v, Data: ChunkData{Size: %#v, %d blocks}}",
					pktTyped.Corner, pktTyped.Data.Size, len(pktTyped.Data.Blocks))
			default:
				p.printf("%#v", pkt)
			}
		}
	}
}
