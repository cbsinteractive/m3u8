package m3u8

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeFromWithOptionsNilMatchesDecodeFrom(t *testing.T) {
	const sample = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000,PATHWAY-ID="fastly"
https://example.com/low.m3u8
`
	r := strings.NewReader(sample)
	p1, lt1, err1 := DecodeFrom(r, false)
	if err1 != nil {
		t.Fatal(err1)
	}
	r2 := strings.NewReader(sample)
	p2, lt2, err2 := DecodeFromWithOptions(r2, false, nil)
	if err2 != nil {
		t.Fatal(err2)
	}
	if lt1 != lt2 {
		t.Fatalf("list type %v vs %v", lt1, lt2)
	}
	m1 := p1.(*MasterPlaylist)
	m2 := p2.(*MasterPlaylist)
	if len(m1.Variants) != len(m2.Variants) {
		t.Fatal("variant count mismatch")
	}
	if m1.Variants[0].PathwayID != m2.Variants[0].PathwayID {
		t.Fatal("pathway mismatch")
	}
}

func TestDecodeFromWithOptionsContentSteeringParsesTagAndPathway(t *testing.T) {
	const sample = `#EXTM3U
#EXT-X-VERSION:9
#EXT-X-CONTENT-STEERING:SERVER-URI="https://steer.example/api",PATHWAY-ID="fastly"
#EXT-X-STREAM-INF:BANDWIDTH=1280000,PATHWAY-ID="fastly"
https://cdn.example/low.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2560000,PATHWAY-ID="akamai"
https://cdn2.example/mid.m3u8
`
	opts := &Options{ContentSteering: &ContentSteeringOptions{}}
	pl, lt, err := DecodeFromWithOptions(strings.NewReader(sample), true, opts)
	if err != nil {
		t.Fatal(err)
	}
	if lt != MASTER {
		t.Fatalf("expected MASTER, got %v", lt)
	}
	m := pl.(*MasterPlaylist)
	if m.ContentSteeringServerURI != "https://steer.example/api" {
		t.Fatalf("SERVER-URI: got %q", m.ContentSteeringServerURI)
	}
	if m.ContentSteeringPathwayID != "fastly" {
		t.Fatalf("default PATHWAY-ID: got %q", m.ContentSteeringPathwayID)
	}
	if len(m.Variants) != 2 {
		t.Fatalf("variants: %d", len(m.Variants))
	}
	if m.Variants[0].PathwayID != "fastly" || m.Variants[1].PathwayID != "akamai" {
		t.Fatalf("variant pathways: %+v, %+v", m.Variants[0].PathwayID, m.Variants[1].PathwayID)
	}
}

func TestDecodeFromWithOptionsWithoutContentSteeringIgnoresPathwayID(t *testing.T) {
	const sample = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000,PATHWAY-ID="fastly"
https://example.com/low.m3u8
`
	pl, _, err := DecodeFromWithOptions(strings.NewReader(sample), false, &Options{})
	if err != nil {
		t.Fatal(err)
	}
	m := pl.(*MasterPlaylist)
	if m.Variants[0].PathwayID != "" {
		t.Fatalf("expected empty PathwayID without ContentSteering sub-option, got %q", m.Variants[0].PathwayID)
	}
}

func TestContentSteeringEncodeRoundTrip(t *testing.T) {
	m := NewMasterPlaylist()
	m.ContentSteeringServerURI = "https://steer.example/v1/steer"
	m.ContentSteeringPathwayID = "fastly"
	m.Append("https://a.example/out/v1/x/manifest.m3u8", nil, VariantParams{
		ProgramId: 1, Bandwidth: 1000000, Codecs: "avc1.64001E,mp4a.40.2", PathwayID: "fastly",
	})
	m.Append("https://b.example/out/v1/x/manifest.m3u8", nil, VariantParams{
		ProgramId: 1, Bandwidth: 1000000, Codecs: "avc1.64001E,mp4a.40.2", PathwayID: "akamai",
	})

	encoded := m.Encode().String()
	opts := &Options{ContentSteering: &ContentSteeringOptions{}}
	pl, _, err := DecodeFromWithOptions(bytes.NewReader([]byte(encoded)), true, opts)
	if err != nil {
		t.Fatalf("decode: %v\n%s", err, encoded)
	}
	m2 := pl.(*MasterPlaylist)
	if m2.ContentSteeringServerURI != m.ContentSteeringServerURI {
		t.Fatalf("server URI %q vs %q", m2.ContentSteeringServerURI, m.ContentSteeringServerURI)
	}
	if m2.ContentSteeringPathwayID != m.ContentSteeringPathwayID {
		t.Fatalf("pathway %q vs %q", m2.ContentSteeringPathwayID, m.ContentSteeringPathwayID)
	}
	if len(m2.Variants) != 2 {
		t.Fatal(len(m2.Variants))
	}
	if m2.Variants[0].PathwayID != "fastly" || m2.Variants[1].PathwayID != "akamai" {
		t.Fatalf("pathways %+v %+v", m2.Variants[0].PathwayID, m2.Variants[1].PathwayID)
	}
}

func TestDecodeWithStillWorks(t *testing.T) {
	const sample = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000
https://example.com/low.m3u8
`
	_, _, err := DecodeWith(strings.NewReader(sample), false, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecodeFromWithOptionsAppliesSteeringOptionsForEncode(t *testing.T) {
	const sample = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000,CODECS="avc1.64001E,mp4a.40.2"
https://origin.example.com/asset/low.m3u8
`
	opts := &Options{ContentSteering: &ContentSteeringOptions{
		ServerURI:        "https://steer.example/steer",
		DefaultPathwayID: "cdn-a",
		Pathways: []HLSSteeringPathway{
			{ID: "cdn-a", BaseURL: "https://cdn-a.example.com"},
			{ID: "cdn-b", BaseURL: "https://cdn-b.example.com"},
		},
	}}
	pl, lt, err := DecodeFromWithOptions(strings.NewReader(sample), true, opts)
	if err != nil {
		t.Fatal(err)
	}
	if lt != MASTER {
		t.Fatalf("list type %v", lt)
	}
	m := pl.(*MasterPlaylist)
	if m.ContentSteeringServerURI != "https://steer.example/steer" {
		t.Fatalf("SERVER-URI %q", m.ContentSteeringServerURI)
	}
	if m.ContentSteeringPathwayID != "cdn-a" {
		t.Fatalf("default pathway %q", m.ContentSteeringPathwayID)
	}
	if len(m.Variants) != 2 {
		t.Fatalf("want 2 pathway variants, got %d", len(m.Variants))
	}
	enc := m.Encode().String()
	if !strings.Contains(enc, "#EXT-X-CONTENT-STEERING:") || !strings.Contains(enc, "https://steer.example/steer") {
		t.Fatalf("missing EXT-X-CONTENT-STEERING in:\n%s", enc)
	}
	if !strings.Contains(enc, "PATHWAY-ID=\"cdn-a\"") || !strings.Contains(enc, "PATHWAY-ID=\"cdn-b\"") {
		t.Fatalf("missing pathway IDs in:\n%s", enc)
	}
	if !strings.Contains(enc, "https://cdn-a.example.com") || !strings.Contains(enc, "https://cdn-b.example.com") {
		t.Fatalf("missing CDN hosts in:\n%s", enc)
	}
}
