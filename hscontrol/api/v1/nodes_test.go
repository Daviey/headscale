package apiv1

import (
	"net/netip"
	"testing"

	"tailscale.com/net/tsaddr"
)

func TestApprovedRoutesFromRequest(t *testing.T) {
	announced := []netip.Prefix{
		mustPrefix(t, "0.0.0.0/0"),
		mustPrefix(t, "::/0"),
		mustPrefix(t, "10.1.0.0/16"),
	}

	for _, tc := range []struct {
		name      string
		requested []string
		want      []string
		wantErr   bool
	}{
		{
			name:      "empty clear",
			requested: nil,
			want:      nil,
		},
		{
			name:      "v4 exit couples families",
			requested: []string{"0.0.0.0/0"},
			want:      []string{"0.0.0.0/0", "::/0"},
		},
		{
			name:      "v6 exit couples families",
			requested: []string{"::/0"},
			want:      []string{"0.0.0.0/0", "::/0"},
		},
		{
			name:      "subnet passthrough",
			requested: []string{"10.1.0.0/16"},
			want:      []string{"10.1.0.0/16"},
		},
		{
			name:      "coupling dedupes",
			requested: []string{"0.0.0.0/0", "::/0"},
			want:      []string{"0.0.0.0/0", "::/0"},
		},
		{
			name:      "unannounced route rejected",
			requested: []string{"10.99.0.0/16"},
			wantErr:   true,
		},
		{
			name:      "unannounced alongside approved rejected",
			requested: []string{"0.0.0.0/0", "192.0.2.0/24"},
			wantErr:   true,
		},
		{
			name:      "unparsable route rejected",
			requested: []string{"not-a-prefix"},
			wantErr:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := approvedRoutesFromRequest(tc.requested, announced)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}

			for i := range got {
				want := mustPrefix(t, tc.want[i])
				if got[i] != want {
					t.Fatalf("route %d: got %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestApprovedRoutesFromRequestNoAnnouncements(t *testing.T) {
	if _, err := approvedRoutesFromRequest([]string{"0.0.0.0/0"}, nil); err == nil {
		t.Fatal("expected unannounced exit route to be rejected for silent node")
	}
}

func mustPrefix(t *testing.T, s string) netip.Prefix {
	t.Helper()

	p, err := netip.ParsePrefix(s)
	if err != nil {
		t.Fatalf("parsing %s: %v", s, err)
	}

	return p
}

var _ = tsaddr.AllIPv4 // keep import stable if cases change
