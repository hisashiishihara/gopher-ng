package conformance

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
)

type vector struct {
	ID          string     `json:"id"`
	Kind        string     `json:"kind"`
	WireHex     string     `json:"wire_hex"`
	URI         string     `json:"uri"`
	Valid       bool       `json:"valid"`
	Selector    string     `json:"selector"`
	Host        string     `json:"host"`
	Port        uint16     `json:"port"`
	Records     [][]string `json:"records"`
	AppliesTo   string     `json:"applies_to"`
	LimitBytes  int64      `json:"limit_bytes"`
	RepeatASCII string     `json:"repeat_ascii"`
	RepeatCount int        `json:"repeat_count"`
	SuffixHex   string     `json:"suffix_hex"`
}

func TestConformance(t *testing.T) {
	for _, name := range []string{"normative", "policies"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("v0.0.1/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			var corpus struct {
				CoreVersion string   `json:"core_version"`
				Vectors     []vector `json:"vectors"`
			}
			if err := json.Unmarshal(data, &corpus); err != nil {
				t.Fatal(err)
			}
			if corpus.CoreVersion != "v0.0.1" || len(corpus.Vectors) == 0 {
				t.Fatal("unexpected or empty corpus")
			}
			for _, v := range corpus.Vectors {
				t.Run(v.ID, func(t *testing.T) {
					if v.AppliesTo != "" && v.AppliesTo != "go" {
						t.Skip("policy for " + v.AppliesTo)
					}
					wire, err := hex.DecodeString(v.WireHex)
					if err != nil {
						t.Fatal(err)
					}
					suffix, err := hex.DecodeString(v.SuffixHex)
					if err != nil {
						t.Fatal(err)
					}
					wire = append(wire, []byte(strings.Repeat(v.RepeatASCII, v.RepeatCount))...)
					wire = append(wire, suffix...)
					var got, want any
					switch v.Kind {
					case "request":
						got, err = protocol.ReadSelector(bytes.NewReader(wire))
						want = v.Selector
					case "uri":
						got, err = protocol.ParseURI(v.URI)
						want = protocol.URI{Host: v.Host, Port: v.Port, Selector: v.Selector}
					case "response":
						var records []protocol.Record
						if v.LimitBytes > 0 {
							records, err = protocol.ParseResponseWithLimit(bytes.NewReader(wire), v.LimitBytes)
						} else {
							records, err = protocol.ParseResponse(bytes.NewReader(wire))
						}
						rows := make([][]string, 0, len(records))
						for _, record := range records {
							rows = append(rows, append([]string{string(record.Type)}, record.Fields...))
						}
						got, want = rows, v.Records
					default:
						t.Fatalf("unknown kind %q", v.Kind)
					}
					if (err == nil) != v.Valid {
						t.Fatalf("valid=%v, error=%v", v.Valid, err)
					}
					if v.Valid && !reflect.DeepEqual(got, want) {
						t.Fatalf("got %#v, want %#v", got, want)
					}
				})
			}
		})
	}
}
