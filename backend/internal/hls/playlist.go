// Package hls parses and rewrites HLS (.m3u8) playlists so that media segments
// and sub-playlists stored in a private bucket can be served via signed URLs,
// without ever exposing the bucket publicly.
//
// The Transcoder API writes a media/rendition playlist as a flat list of
// relative segment references (plus an #EXT-X-MAP init segment). Those relative
// references cannot be signed, so the API serves the playlist and rewrites each
// reference into an absolute signed URL before handing it to the player.
package hls

import "strings"

// MasterPlaylistName is the file name of the HLS master playlist emitted by the
// Transcoder template (see internal/config/transcoder/hls_adaptive.json). It is
// the single source of truth shared by the processor (which records the
// playback URL) and the delivery handler (which serves it as a passthrough).
const MasterPlaylistName = "manifest.m3u8"

// URIResolver maps a URI reference found in a playlist (for example
// "video-1080p0.m4s" or "video-1080p.m3u8") to the URL that should replace it,
// such as a signed GCS download URL. Returning the input unchanged leaves the
// reference as-is; deciding whether to rewrite (e.g. skipping already-absolute
// URLs) is the resolver's responsibility.
//
// It is called for every non-empty URI reference in a playlist: bare
// segment/sub-playlist lines, and the URI="..." attribute of tags such as
// #EXT-X-MAP, #EXT-X-MEDIA and #EXT-X-I-FRAME-STREAM-INF.
type URIResolver func(uri string) (string, error)

// RewritePlaylist rewrites every URI reference in an m3u8 playlist using
// resolve. Blank lines and tag/comment lines without a URI attribute are
// preserved verbatim, and the playlist's line structure (including CRLF
// endings) round-trips unchanged. The first resolve error aborts and is
// returned.
func RewritePlaylist(content string, resolve URIResolver) (string, error) {
	lines := strings.Split(content, "\n")

	for i, raw := range lines {
		// Preserve a trailing CR so CRLF playlists round-trip byte-for-byte.
		line := raw
		cr := ""
		if strings.HasSuffix(line, "\r") {
			cr = "\r"
			line = strings.TrimSuffix(line, "\r")
		}

		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			// Blank line: leave as-is.
		case strings.HasPrefix(trimmed, "#"):
			// Tag line: rewrite a URI="..." attribute if one is present.
			rewritten, err := rewriteURIAttribute(line, resolve)
			if err != nil {
				return "", err
			}
			line = rewritten
		default:
			// Bare URI line: a segment or sub-playlist reference.
			resolved, err := resolve(trimmed)
			if err != nil {
				return "", err
			}
			line = resolved
		}

		lines[i] = line + cr
	}

	return strings.Join(lines, "\n"), nil
}

// rewriteURIAttribute replaces the value of a single URI="..." attribute on a
// tag line. HLS tags carry at most one such attribute. Lines without a
// well-formed URI attribute are returned unchanged.
func rewriteURIAttribute(line string, resolve URIResolver) (string, error) {
	const marker = `URI="`

	idx := strings.Index(line, marker)
	if idx == -1 {
		return line, nil
	}

	start := idx + len(marker)
	rel := strings.Index(line[start:], `"`)
	if rel == -1 {
		return line, nil // Unterminated attribute; leave the line untouched.
	}
	end := start + rel

	uri := line[start:end]
	if uri == "" {
		return line, nil
	}

	resolved, err := resolve(uri)
	if err != nil {
		return "", err
	}

	return line[:start] + resolved + line[end:], nil
}
