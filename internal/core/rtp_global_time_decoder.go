package core

import (
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v3/pkg/rtptime"

	"github.com/pion/rtp"
)

type rtpGlobalTimeDecoderTrack struct {
	startPTS time.Duration
	decoder  *rtptime.Decoder
}

type rtpGlobalTimeDecoder struct {
	mutex          sync.Mutex
	startNTPFilled bool
	startNTP       time.Time
	tracks         map[interface{}]*rtpGlobalTimeDecoderTrack
}

func newRTSPTimeDecoder() *rtpGlobalTimeDecoder {
	return &rtpGlobalTimeDecoder{
		tracks: make(map[interface{}]*rtpGlobalTimeDecoderTrack),
	}
}

func (d *rtpGlobalTimeDecoder) decode(
	track interface{},
	clockRate int,
	ptsEqualsDTS bool,
	pkt *rtp.Packet,
) (time.Duration, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	df, ok := d.tracks[track]

	if !ok {
		if !ptsEqualsDTS {
			return 0, false
		}

		now := time.Now()

		if !d.startNTPFilled {
			d.startNTPFilled = true
			d.startNTP = now
		}

		df = &rtpGlobalTimeDecoderTrack{
			startPTS: now.Sub(d.startNTP),
			decoder:  rtptime.NewDecoder(clockRate),
		}

		d.tracks[track] = df
		df.decoder.Decode(pkt.Timestamp)

		return df.startPTS, true
	}

	return df.startPTS + df.decoder.Decode(pkt.Timestamp), true
}
