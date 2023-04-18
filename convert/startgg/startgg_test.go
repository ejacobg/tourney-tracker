package startgg

import (
	"github.com/ejacobg/tourney-tracker/convert"
	"golang.org/x/exp/slices"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func routes() {
	http.HandleFunc("/no-reset", serveFile("no-reset.json"))
	http.HandleFunc("/reset-no-points", serveFile("reset-no-points.json"))
	http.HandleFunc("/reset-with-points", serveFile("reset-with-points.json"))
}

func serveFile(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "testdata/"+file)
	}
}

func Test_response(t *testing.T) {
	tests := []struct {
		name           string
		tournamentName string
		tournamentURL  string
		bracketReset   bool
		placements     []int64
		numEntrants    int
	}{
		{"no-reset", "Silver State Smash x Pirate Hackers Black Lives Matter Charity Tournament - Singles 1v1", "https://start.gg/tournament/silver-state-smash-x-pirate-hackers-black-lives-matter-charity/event/singles-1v1", false, []int64{33, 25, 17, 13, 9, 7, 5, 4, 3, 2, 1}, 42},
		{"reset-no-points", "Wrangler Rumble #1 - Ultimate Singles", "https://start.gg/tournament/wrangler-rumble-1/event/ultimate-singles", false, []int64{13, 9, 7, 5, 4, 3, 2, 1}, 13},
		{"reset-with-points", "Shinto Series: Smash #1 - Singles 1v1", "https://start.gg/tournament/shinto-series-smash-1/event/singles-1v1", true, []int64{97, 65, 49, 33, 25, 17, 13, 9, 7, 5, 4, 3, 2, 1}, 128},
	}

	// Attach our routes to the DefaultServeMux.
	routes()
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We're making our own request rather than using newRequest(). See Test_newRequest() for tests involving that function.
			req, err := http.NewRequest(http.MethodGet, ts.URL+"/"+tt.name, nil)
			if err != nil {
				t.Error("failed to create request:", err)
				return
			}
			res, err := convert.Get[response](req)
			if err != nil {
				t.Error("failed to get tournament:", err)
				return
			}
			tourney := res.tournament()
			if tourney.Name != tt.tournamentName {
				t.Errorf("response tournamentName = %v, want %v", tourney.Name, tt.tournamentName)
			}
			if tourney.URL != tt.tournamentURL {
				t.Errorf("response tournamentURL = %v, want %v", tourney.URL, tt.tournamentURL)
			}
			if tourney.BracketReset != tt.bracketReset {
				t.Errorf("applyResetPoints() BracketReset = %v, want %v", tourney.BracketReset, tt.bracketReset)
			}
			if !slices.Equal(tourney.Placements, tt.placements) {
				t.Errorf("uniquePlacements() Placements = %v, want %v", tourney.Placements, tt.placements)
			}
			entrants := res.entrants()
			if len(entrants) != tt.numEntrants {
				t.Errorf("entrants() length = %v, want %v", len(entrants), tt.numEntrants)
			}
		})
	}
}

func Test_parseSlugs(t *testing.T) {
	type args struct {
		URL *url.URL
	}
	tests := []struct {
		name               string
		args               args
		wantTournamentSlug string
		wantEventSlug      string
		wantErr            bool
	}{
		{"correct path", args{&url.URL{Path: "/tournament/shinto-series-smash-1/event/singles-1v1"}}, "shinto-series-smash-1", "singles-1v1", false},
		{"path too short", args{&url.URL{Path: "/tournament/shinto-series-smash-1"}}, "", "", true},
		{"path too long", args{&url.URL{Path: "/tournament/shinto-series-smash-1/event/singles-1v1/standings"}}, "shinto-series-smash-1", "singles-1v1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTournamentSlug, gotEventSlug, err := parseSlugs(tt.args.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSlugs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTournamentSlug != tt.wantTournamentSlug {
				t.Errorf("parseSlugs() gotTournamentSlug = %v, want %v", gotTournamentSlug, tt.wantTournamentSlug)
			}
			if gotEventSlug != tt.wantEventSlug {
				t.Errorf("parseSlugs() gotEventSlug = %v, want %v", gotEventSlug, tt.wantEventSlug)
			}
		})
	}
}

func Test_newRequest(t *testing.T) {
	type args struct {
		tournamentSlug string
		eventSlug      string
		key            string
	}
	tests := []struct {
		name    string
		args    args
		want    http.Header
		wantErr bool
	}{
		// Only confirming that the headers are correct.
		{"correct headers", args{"shinto-series-smash-1", "singles-1v1", "api-key"}, http.Header{"Content-Type": {"application/json"}, "Authorization": {"Bearer api-key"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRequest(tt.args.tournamentSlug, tt.args.eventSlug, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("newRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Header, tt.want) {
				t.Errorf("newRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
