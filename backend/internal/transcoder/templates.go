package transcoder

var (
	Resolution1080p = Resolution{
		Name:         "1080p",
		Width:        1920,
		Height:       1080,
		VideoBitrate: 5000000,
		AudioBitrate: 128000,
		FrameRate:    30,
	}

	Resolution720p = Resolution{
		Name:         "720p",
		Width:        1280,
		Height:       720,
		VideoBitrate: 2500000,
		AudioBitrate: 128000,
		FrameRate:    30,
	}

	Resolution480p = Resolution{
		Name:         "480p",
		Width:        854,
		Height:       480,
		VideoBitrate: 1000000,
		AudioBitrate: 128000,
		FrameRate:    30,
	}
)

func GetStandardResolutions() []Resolution {
	return []Resolution{
		Resolution1080p,
		Resolution720p,
		Resolution480p,
	}
}
