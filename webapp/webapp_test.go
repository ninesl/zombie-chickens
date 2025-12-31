package webapp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ninesl/zombie-chickens/webapp/router"
	"github.com/ninesl/zombie-chickens/webapp/router/endpoints"
	"github.com/ninesl/zombie-chickens/webapp/state"
)

func setupTestRouter() *chi.Mux {
	state.ResetSession()
	r := chi.NewRouter()
	router.GetRoutes(r)
	return r
}

func TestLobbyPageLoads(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Zombie Chickens") {
		t.Error("expected page to contain 'Zombie Chickens'")
	}
	if !strings.Contains(body, `id="lobby-content"`) {
		t.Error("expected page to contain lobby-content div")
	}
	if !strings.Contains(body, "Join Game") {
		t.Error("expected page to contain join form")
	}
}

func TestLobbyJoin(t *testing.T) {
	r := setupTestRouter()

	// First request to get session cookie
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Extract session cookie
	cookies := w1.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie to be set")
	}

	// Join lobby
	form := url.Values{}
	form.Set(endpoints.FieldPlayerName, "TestPlayer")
	req2 := httptest.NewRequest("POST", endpoints.LobbyJoin, strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	body := w2.Body.String()
	t.Logf("Join response: %s", body)

	// Response should contain the wrapper div for HTMX swap
	if !strings.Contains(body, `id="lobby-content"`) {
		t.Error("expected response to contain lobby-content wrapper div")
	}
	if !strings.Contains(body, "TestPlayer") {
		t.Error("expected response to contain player name")
	}
	if !strings.Contains(body, "Start Game") {
		t.Error("expected first player to see Start Game button")
	}
}

func TestLobbyStart(t *testing.T) {
	r := setupTestRouter()

	// Get session and join
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	cookies := w1.Result().Cookies()

	form := url.Values{}
	form.Set(endpoints.FieldPlayerName, "Player1")
	req2 := httptest.NewRequest("POST", endpoints.LobbyJoin, strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// Start game
	req3 := httptest.NewRequest("POST", endpoints.LobbyStart, nil)
	for _, c := range cookies {
		req3.AddCookie(c)
	}
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w3.Code)
	}

	body := w3.Body.String()
	t.Logf("Start response: %s", body)

	if !strings.Contains(body, "Game has started") {
		t.Error("expected response to indicate game started")
	}

	// Verify game state
	session := state.GetSession()
	if !session.IsStarted() {
		t.Error("expected game to be started")
	}
}

func TestGamePageRequiresStartedGame(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest("GET", endpoints.GamePage, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should redirect to lobby since game not started
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected redirect (303), got %d", w.Code)
	}
}

func TestGameInput(t *testing.T) {
	r := setupTestRouter()

	// Setup: join and start game
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	cookies := w1.Result().Cookies()

	form := url.Values{}
	form.Set(endpoints.FieldPlayerName, "Player1")
	req2 := httptest.NewRequest("POST", endpoints.LobbyJoin, strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	req3 := httptest.NewRequest("POST", endpoints.LobbyStart, nil)
	for _, c := range cookies {
		req3.AddCookie(c)
	}
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	// Now submit game input (choice 0 = skip discard)
	inputForm := url.Values{}
	inputForm.Set(endpoints.FieldChoice, "0")
	req4 := httptest.NewRequest("POST", endpoints.GameInput, strings.NewReader(inputForm.Encode()))
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req4.AddCookie(c)
	}
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w4.Code)
	}

	// Verify game advanced
	session := state.GetSession()
	pending := session.PendingInput()
	if pending == nil {
		t.Error("expected pending input after skipping discard")
	} else {
		t.Logf("Pending input: %s", pending.Message)
	}
}

func TestGameBoardRendering(t *testing.T) {
	r := setupTestRouter()

	// Setup: join and start game
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	cookies := w1.Result().Cookies()

	form := url.Values{}
	form.Set(endpoints.FieldPlayerName, "Player1")
	req2 := httptest.NewRequest("POST", endpoints.LobbyJoin, strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	req3 := httptest.NewRequest("POST", endpoints.LobbyStart, nil)
	for _, c := range cookies {
		req3.AddCookie(c)
	}
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	// Render game board directly via components
	session := state.GetSession()
	boardHTML := string(state.GetSession().Game().CurrentPlayer().Name())
	t.Logf("Current player: %s", boardHTML)

	// Check pending input
	pending := session.PendingInput()
	if pending == nil {
		t.Fatal("expected pending input")
	}
	t.Logf("Pending: %s, choices: %v", pending.Message, pending.ValidChoices)

	// Verify game state
	game := session.Game()
	if game.PlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", game.PlayerCount())
	}
	if game.CurrentPlayerIdx() != 0 {
		t.Errorf("expected current player idx 0, got %d", game.CurrentPlayerIdx())
	}
}

func TestMultiplePlayers(t *testing.T) {
	r := setupTestRouter()

	// Player 1 joins
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	cookies1 := w1.Result().Cookies()

	form1 := url.Values{}
	form1.Set(endpoints.FieldPlayerName, "Alice")
	req1j := httptest.NewRequest("POST", endpoints.LobbyJoin, strings.NewReader(form1.Encode()))
	req1j.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies1 {
		req1j.AddCookie(c)
	}
	w1j := httptest.NewRecorder()
	r.ServeHTTP(w1j, req1j)

	// Player 2 joins (different session)
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	cookies2 := w2.Result().Cookies()

	form2 := url.Values{}
	form2.Set(endpoints.FieldPlayerName, "Bob")
	req2j := httptest.NewRequest("POST", endpoints.LobbyJoin, strings.NewReader(form2.Encode()))
	req2j.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies2 {
		req2j.AddCookie(c)
	}
	w2j := httptest.NewRecorder()
	r.ServeHTTP(w2j, req2j)

	// Verify both players in session
	session := state.GetSession()
	if session.PlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", session.PlayerCount())
	}

	players := session.Players()
	if players[0].Name != "Alice" || players[1].Name != "Bob" {
		t.Errorf("unexpected player names: %v", players)
	}

	// Player 2 should NOT see start button (only player 1 can start)
	body2 := w2j.Body.String()
	if strings.Contains(body2, "Start Game") {
		t.Error("player 2 should not see Start Game button")
	}

	// Player 1's view should have start button
	// Re-fetch player 1's view
	req1v := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookies1 {
		req1v.AddCookie(c)
	}
	w1v := httptest.NewRecorder()
	r.ServeHTTP(w1v, req1v)
	body1 := w1v.Body.String()
	if !strings.Contains(body1, "Start Game") {
		t.Error("player 1 should see Start Game button")
	}
}
