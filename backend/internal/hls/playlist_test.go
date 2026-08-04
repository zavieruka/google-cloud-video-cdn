package hls_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zavieruka/video-platform/backend/internal/hls"
)

// signResolver mimics the production resolver: it rewrites relative references
// to a fake signed URL and leaves already-absolute URLs untouched.
func signResolver(uri string) (string, error) {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri, nil
	}
	return "https://signed.example/" + uri + "?sig=abc", nil
}

func TestRewritePlaylist_MediaPlaylist_SignsSegmentsAndInit(t *testing.T) {
	input := strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-VERSION:7",
		"#EXT-X-TARGETDURATION:6",
		"#EXT-X-MEDIA-SEQUENCE:0",
		"#EXT-X-PLAYLIST-TYPE:VOD",
		`#EXT-X-MAP:URI="video-1080pinit.mp4"`,
		"#EXTINF:6.000,",
		"video-1080p0.m4s",
		"#EXTINF:6.000,",
		"video-1080p1.m4s",
		"#EXT-X-ENDLIST",
		"",
	}, "\n")

	out, err := hls.RewritePlaylist(input, signResolver)
	require.NoError(t, err)

	// Init segment (URI attribute) and both media segments are signed.
	assert.Contains(t, out, `#EXT-X-MAP:URI="https://signed.example/video-1080pinit.mp4?sig=abc"`)
	assert.Contains(t, out, "https://signed.example/video-1080p0.m4s?sig=abc")
	assert.Contains(t, out, "https://signed.example/video-1080p1.m4s?sig=abc")

	// Original relative references are gone.
	assert.NotContains(t, out, "\nvideo-1080p0.m4s\n")
	assert.NotContains(t, out, `URI="video-1080pinit.mp4"`)

	// Structural tags are untouched.
	assert.Contains(t, out, "#EXTINF:6.000,")
	assert.Contains(t, out, "#EXT-X-ENDLIST")
	assert.Contains(t, out, "#EXT-X-TARGETDURATION:6")
}

func TestRewritePlaylist_MasterPlaylist_RewritesMediaAndStreamRefs(t *testing.T) {
	input := strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-VERSION:7",
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",AUTOSELECT=YES,URI="audio.m3u8"`,
		`#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080,AUDIO="audio"`,
		"video-1080p.m3u8",
		`#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720,AUDIO="audio"`,
		"video-720p.m3u8",
		"",
	}, "\n")

	out, err := hls.RewritePlaylist(input, signResolver)
	require.NoError(t, err)

	assert.Contains(t, out, `URI="https://signed.example/audio.m3u8?sig=abc"`)
	assert.Contains(t, out, "https://signed.example/video-1080p.m3u8?sig=abc")
	assert.Contains(t, out, "https://signed.example/video-720p.m3u8?sig=abc")
	// The GROUP-ID / AUDIO attributes must not be mistaken for URIs.
	assert.Contains(t, out, `GROUP-ID="audio"`)
	assert.Contains(t, out, `AUDIO="audio"`)
}

func TestRewritePlaylist_PreservesAbsoluteURLsViaResolver(t *testing.T) {
	input := strings.Join([]string{
		"#EXTM3U",
		"#EXTINF:6.000,",
		"https://cdn.example/already-absolute.m4s",
		"",
	}, "\n")

	out, err := hls.RewritePlaylist(input, signResolver)
	require.NoError(t, err)

	assert.Contains(t, out, "https://cdn.example/already-absolute.m4s")
	assert.NotContains(t, out, "signed.example")
}

func TestRewritePlaylist_PreservesCRLFAndBlankLines(t *testing.T) {
	input := "#EXTM3U\r\n#EXTINF:6.000,\r\nseg0.m4s\r\n\r\n#EXT-X-ENDLIST\r\n"

	out, err := hls.RewritePlaylist(input, signResolver)
	require.NoError(t, err)

	// CRLF endings and the blank line survive.
	assert.Contains(t, out, "#EXTM3U\r\n")
	assert.Contains(t, out, "https://signed.example/seg0.m4s?sig=abc\r\n")
	assert.Contains(t, out, "\r\n\r\n") // blank line preserved
	assert.True(t, strings.HasSuffix(out, "#EXT-X-ENDLIST\r\n"))
}

func TestRewritePlaylist_TagWithoutURIUnchanged(t *testing.T) {
	input := "#EXT-X-INDEPENDENT-SEGMENTS\n#EXT-X-START:TIME-OFFSET=0\n"

	out, err := hls.RewritePlaylist(input, signResolver)
	require.NoError(t, err)

	assert.Equal(t, input, out)
}

func TestRewritePlaylist_ResolverErrorPropagates(t *testing.T) {
	boom := errors.New("sign failed")
	resolver := func(string) (string, error) { return "", boom }

	_, err := hls.RewritePlaylist("#EXTM3U\nseg0.m4s\n", resolver)
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

func TestRewritePlaylist_UnterminatedURIAttributeLeftUntouched(t *testing.T) {
	input := `#EXT-X-MAP:URI="broken.mp4` + "\n"

	out, err := hls.RewritePlaylist(input, signResolver)
	require.NoError(t, err)
	assert.Equal(t, input, out)
}
