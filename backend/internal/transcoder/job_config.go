package transcoder

import (
	"fmt"
	"time"

	transcoderpb "cloud.google.com/go/video/transcoder/apiv1/transcoderpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

type JobConfig struct {
	InputURI        string
	OutputURIPrefix string
	Resolutions     []Resolution
}

type Resolution struct {
	Name         string
	Width        int32
	Height       int32
	VideoBitrate int32
	AudioBitrate int32
	FrameRate    int32
}

func NewStandardJobConfig(inputURI, outputPrefix string) *JobConfig {
	return &JobConfig{
		InputURI:        inputURI,
		OutputURIPrefix: outputPrefix,
		Resolutions:     GetStandardResolutions(),
	}
}

func (c *Client) buildJobRequest(config *JobConfig) *transcoderpb.CreateJobRequest {
	elementaryStreams := []*transcoderpb.ElementaryStream{}
	muxStreams := []*transcoderpb.MuxStream{}
	manifestKeys := []string{}

	for _, res := range config.Resolutions {
		videoStreamKey := fmt.Sprintf("video-%s", res.Name)

		elementaryStreams = append(elementaryStreams, &transcoderpb.ElementaryStream{
			Key: videoStreamKey,
			ElementaryStream: &transcoderpb.ElementaryStream_VideoStream{
				VideoStream: &transcoderpb.VideoStream{
					CodecSettings: &transcoderpb.VideoStream_H264{
						H264: &transcoderpb.VideoStream_H264CodecSettings{
							WidthPixels:     res.Width,
							HeightPixels:    res.Height,
							FrameRate:       float64(res.FrameRate),
							BitrateBps:      res.VideoBitrate,
							PixelFormat:     "yuv420p",
							RateControlMode: "vbr",
							CrfLevel:        21,
							GopMode: &transcoderpb.VideoStream_H264CodecSettings_GopFrameCount{
								GopFrameCount: res.FrameRate * 2,
							},
							Preset: "veryfast",
						},
					},
				},
			},
		})

		muxStreams = append(muxStreams, &transcoderpb.MuxStream{
			Key:               res.Name,
			Container:         "fmp4",
			ElementaryStreams: []string{videoStreamKey, "audio"},
			SegmentSettings: &transcoderpb.SegmentSettings{
				SegmentDuration: durationpb.New(6 * time.Second),
			},
		})

		manifestKeys = append(manifestKeys, res.Name)
	}

	elementaryStreams = append(elementaryStreams, &transcoderpb.ElementaryStream{
		Key: "audio",
		ElementaryStream: &transcoderpb.ElementaryStream_AudioStream{
			AudioStream: &transcoderpb.AudioStream{
				Codec:           "aac",
				BitrateBps:      128000,
				ChannelCount:    2,
				SampleRateHertz: 48000,
			},
		},
	})

	manifests := []*transcoderpb.Manifest{
		{
			FileName:   "manifest.m3u8",
			Type:       transcoderpb.Manifest_HLS,
			MuxStreams: manifestKeys,
		},
	}

	return &transcoderpb.CreateJobRequest{
		Parent: c.getParent(),
		Job: &transcoderpb.Job{
			InputUri:  config.InputURI,
			OutputUri: config.OutputURIPrefix,
			JobConfig: &transcoderpb.Job_Config{
				Config: &transcoderpb.JobConfig{
					ElementaryStreams: elementaryStreams,
					MuxStreams:        muxStreams,
					Manifests:         manifests,
				},
			},
		},
	}
}
