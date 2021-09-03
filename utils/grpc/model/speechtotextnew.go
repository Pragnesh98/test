package model

import (
	"context"
	"io"

	pb "bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/proto"
)

type SpeechToTextNew interface {
	STTStreamingNew(context.Context, io.Reader) (chan *pb.RecognizeResponse, error)
	Close()
}
