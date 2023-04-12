package challonge

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
		{"no-reset", "(SSC C TIER) Gator Grind #9", "https://challonge.com/kpqlgghc", false, []int64{17, 13, 9, 7, 5, 4, 3, 2, 1}, 20},
		{"reset-no-points", "(SSC C Tier) Gator Grind #12", "https://challonge.com/8ozc6ffz", false, []int64{17, 13, 9, 7, 5, 4, 3, 2, 1}, 20},
		{"reset-with-points", "(SSC C Tier) Gator Grind #7", "https://challonge.com/t4kq4f5b", true, []int64{17, 13, 9, 7, 5, 4, 3, 2, 1}, 24},
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
				t.Error("failed to get http:", err)
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

func Test_parseID(t *testing.T) {
	type args struct {
		URL *url.URL
	}
	tests := []struct {
		name             string
		args             args
		wantTournamentID string
		wantErr          bool
	}{
		{"correct path", args{&url.URL{Path: "/correct"}}, "correct", false},
		{"path too short", args{&url.URL{Path: ""}}, "", true},
		{"path too long", args{&url.URL{Path: "/too/long"}}, "too", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTournamentID, err := parseID(tt.args.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTournamentID != tt.wantTournamentID {
				t.Errorf("parseID() gotTournamentID = %v, want %v", gotTournamentID, tt.wantTournamentID)
			}
		})
	}
}

func Test_newRequest(t *testing.T) {
	type args struct {
		tournamentID string
		username     string
		password     string
	}
	tests := []struct {
		name       string
		args       args
		wantURL    string
		wantHeader http.Header
		wantErr    bool
	}{
		{"correct url and headers", args{"test", "username", "password"}, apiURL + "test.json?include_matches=1&include_participants=1", http.Header{"Authorization": {"Basic dXNlcm5hbWU6cGFzc3dvcmQ="}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRequest(tt.args.tournamentID, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("newRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.URL.String() != tt.wantURL {
				t.Errorf("newRequest() url = %v, want %v", got.URL.String(), tt.wantURL)
			}
			if !reflect.DeepEqual(got.Header, tt.wantHeader) {
				t.Errorf("newRequest() got = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}
