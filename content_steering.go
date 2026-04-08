package m3u8

import (
	"net/url"
	"strings"
)

// ApplyContentSteeringOptions merges o into p so Encode outputs policy-driven steering
// (RFC 8216 bis §4.4.6.6, §4.4.6.2): EXT-X-CONTENT-STEERING and variant PATHWAY-ID / URIs.
//
// Non-empty ServerURI and DefaultPathwayID override p's tag fields. When Pathways is
// non-empty, each existing variant is expanded to len(Pathways) copies (one per pathway);
// each copy gets Pathway-ID from the pathway ID and, if BaseURL is set, a URI derived
// via joinSteeringVariantURIBase.
//
// DecodeFromWithOptions calls this after a successful master decode when opts.ContentSteering is non-nil.
func ApplyContentSteeringOptions(p *MasterPlaylist, o *ContentSteeringOptions) {
	if p == nil || o == nil {
		return
	}
	if o.ServerURI != "" {
		p.ContentSteeringServerURI = string(o.ServerURI)
	}
	if o.DefaultPathwayID != "" {
		p.ContentSteeringPathwayID = string(o.DefaultPathwayID)
	}
	if len(o.Pathways) == 0 {
		return
	}
	orig := append([]*Variant(nil), p.Variants...)
	p.Variants = nil
	for _, v := range orig {
		if v == nil {
			continue
		}
		for _, pw := range o.Pathways {
			if pw.ID == "" && pw.BaseURL == "" {
				continue
			}
			nv := *v
			if pw.ID != "" {
				nv.PathwayID = string(pw.ID)
			}
			if pw.BaseURL != "" {
				nv.URI = joinSteeringVariantURIBase(pw.BaseURL, v.URI)
			}
			p.Variants = append(p.Variants, &nv)
		}
	}
}

// joinSteeringVariantURIBase builds a variant URI for a pathway: for an absolute variant URI,
// scheme/host (and optional userinfo) come from base; path and query come from the variant.
// If base includes a path prefix, it is prepended before the variant path. Relative variant URIs
// are resolved against base.
func joinSteeringVariantURIBase(base, variantURI string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return variantURI
	}
	b, err := url.Parse(base)
	if err != nil || b.Host == "" {
		return variantURI
	}
	u, err := url.Parse(variantURI)
	if err != nil {
		return variantURI
	}
	if !u.IsAbs() {
		b2, err := url.Parse(strings.TrimSuffix(base, "/") + "/")
		if err != nil {
			return variantURI
		}
		return b2.ResolveReference(u).String()
	}
	out := *u
	out.Scheme = b.Scheme
	out.Host = b.Host
	if b.User != nil {
		out.User = b.User
	}
	if pfx := strings.TrimRight(b.Path, "/"); pfx != "" {
		if strings.HasPrefix(out.Path, "/") {
			out.Path = pfx + out.Path
		} else {
			out.Path = pfx + "/" + out.Path
		}
	}
	return out.String()
}
