package m3u8

// Options configures decode/encode behavior for playlists that use optional extensions.
type Options struct {
	// ContentSteering, when non-nil, enables parsing and writing of RFC 8216 bis
	// content steering tags and attributes (§4.4.6.6, §4.4.6.2).
	ContentSteering *ContentSteeringOptions
}

// PathwayID identifies an HLS content steering pathway (RFC 8216 bis §7.2).
type PathwayID string

// HLSSteeringPathway pairs a pathway identifier with the absolute CDN base URI
// used when emitting duplicate EXT-X-STREAM-INF lines (one set per pathway).
type HLSSteeringPathway struct {
	ID PathwayID

	// BaseURL is the absolute base for variant and media playlist URIs on this pathway.
	BaseURL string
}

// ContentSteeringOptions holds header- or policy-derived values used when rendering
// or rewriting HLS multivariant playlists per RFC 8216 bis §4.4.6.6 (EXT-X-CONTENT-STEERING)
// and §4.4.6.2 (PATHWAY-ID on EXT-X-STREAM-INF).
type ContentSteeringOptions struct {
	// ServerURI is SERVER-URI: quoted URI to the steering manifest (§4.4.6.6). Required
	// when emitting #EXT-X-CONTENT-STEERING.
	ServerURI string

	// DefaultPathwayID is PATHWAY-ID on #EXT-X-CONTENT-STEERING: pathway applied until the
	// first steering manifest is obtained (§4.4.6.6). Optional; must match a PATHWAY-ID
	// on at least one variant when set.
	DefaultPathwayID PathwayID

	// Pathways lists every pathway for PATHWAY-ID on duplicated variant lines and for
	// base-URL selection when rewriting URIs per pathway.
	Pathways []HLSSteeringPathway
}
