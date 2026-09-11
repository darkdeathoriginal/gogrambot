package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/config"
	"github.com/darkdeathoriginal/gogrambot/handler"
	"github.com/darkdeathoriginal/gogrambot/helpers"
	"github.com/darkdeathoriginal/gogrambot/models"
	_ "github.com/darkdeathoriginal/gogrambot/plugins"
	"github.com/joho/godotenv"
)

// --- Global State ---
type LoginState string

const (
	StateStarting  LoginState = "STARTING"
	StateIdle      LoginState = "IDLE"
	StateQR        LoginState = "QR_SCAN"
	StatePassword  LoginState = "PASSWORD_REQUIRED" // 2FA Needed
	StateWrongPass LoginState = "PASSWORD_WRONG"    // 2FA Wrong
	StateLoggedIn  LoginState = "LOGGED_IN"
	StateFailed    LoginState = "FAILED"
)

var (
	client       *telegram.Client
	currentState = StateStarting
	currentUser  *telegram.UserObj
	activeClient *telegram.Client
	clientCtx    context.Context
	cancelClient context.CancelFunc
	qrURL        = ""
	passwordChan = make(chan string) // The bridge between Browser and Callback
	stateMu      sync.Mutex
)

const SessionFile = "telegram.session"

func main() {
	godotenv.Load()
	models.InitDatabase()

	// 1. Init Client
	initClient()

	// 2. HTTP Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// API Endpoints
	http.HandleFunc("/api/poll", pollHandler)               // Frontend checks status here
	http.HandleFunc("/api/login", startLoginHandler)        // Triggers QR generation
	http.HandleFunc("/api/password", submitPasswordHandler) // Receives 2FA from browser
	http.HandleFunc("/api/logout", logoutHandler)

	fmt.Println("Server running at http://localhost:" + config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, basicAuthMiddleware(http.DefaultServeMux)))
}

func initClient() {
	stateMu.Lock()
	if cancelClient != nil {
		cancelClient()
	}
	clientCtx, cancelClient = context.WithCancel(context.Background())
	ctx := clientCtx
	currentState = StateStarting
	currentUser = nil
	qrURL = ""
	stateMu.Unlock()
	appIDStr := config.AppID
	appHash := config.AppHash
	appID, _ := strconv.Atoi(appIDStr)

	c, err := telegram.NewClient(telegram.ClientConfig{
		AppID:        int32(appID),
		AppHash:      appHash,
		Session:      SessionFile, // Saves login to file
		FloodHandler: helpers.TelegramRequests.HandleFlood,
	})

	if err != nil {
		log.Fatal("Client init failed:", err)
	}

	stateMu.Lock()
	client = c
	stateMu.Unlock()
	// Serve the dashboard while a real startup RPC waits for Telegram.
	go func() {
		if err := helpers.RetryTelegramFlood(ctx, "startup connection", c.Connect); err != nil {
			log.Println("Connect error:", err)
			setClientState(c, StateFailed)
			return
		}
		loadClientSession(ctx, c)
	}()
}

func loadClientSession(ctx context.Context, c *telegram.Client) {
	var me *telegram.UserObj
	err := helpers.RetryTelegramFlood(ctx, "startup account check", func() error {
		var err error
		me, err = c.GetMe()
		return err
	})
	if err != nil {
		log.Println("Account check failed:", err)
		if c.MatchRPCError(err, "AUTH_KEY_UNREGISTERED") || c.MatchRPCError(err, "SESSION_REVOKED") || c.MatchRPCError(err, "SESSION_EXPIRED") {
			setClientState(c, StateIdle)
		} else {
			setClientState(c, StateFailed)
		}
		return
	}
	if ctx.Err() != nil {
		return
	}
	stateMu.Lock()
	if client != c || activeClient == c {
		stateMu.Unlock()
		return
	}
	activeClient = c
	stateMu.Unlock()
	c.SetCommandPrefixes(config.CommandPrefix)
	c.SetParseMode(telegram.MarkDown)
	selfFilter := telegram.Any(telegram.FilterOutgoing, telegram.FromUser(me.ID))
	for _, plugin := range handler.Plugins {
		if plugin.OnStart != nil {
			go plugin.OnStart(c)
		}

		if plugin.Handler != nil {
			event := plugin.On
			if event == "" {
				event = "cmd:" + plugin.Name
			}

			var finalFilter telegram.Filter

			switch {
			// Case 1: No custom filter, NOT AllowAll → block self
			case plugin.Filter == nil && !plugin.AllowAll:
				finalFilter = selfFilter

			// Case 2: Custom filter exists, NOT AllowAll → AND with self filter
			case plugin.Filter != nil && !plugin.AllowAll:
				finalFilter = telegram.All(*plugin.Filter, selfFilter)

			// Case 3: Custom filter exists, AllowAll → use as-is
			case plugin.Filter != nil && plugin.AllowAll:
				finalFilter = *plugin.Filter

			// Case 4: No filter + AllowAll → allow everything
			default:
				finalFilter = telegram.Filter{}
			}

			c.On(event, plugin.Handler, finalFilter)
		}
	}
	c.On("cmd:anything", func(m *telegram.NewMessage) error {
		m.Reply("You said: " + m.Text())
		return nil
	}, telegram.FilterOutgoing)
	stateMu.Lock()
	if client == c {
		currentUser = me
		currentState = StateLoggedIn
	}
	stateMu.Unlock()
	log.Printf("Telegram ready: plugins registered for user %d", me.ID)
	// A greeting is optional and must never gate plugin registration or HTTP.
	go func() {
		if err := helpers.TelegramRequests.Do(ctx, helpers.TelegramSendInterval, func() error {
			_, err := c.SendMessage("me", fmt.Sprintf("Hello, %s!", me.FirstName))
			return err
		}); err != nil {
			log.Println("Startup greeting failed:", err)
		}
	}()
}

func setClientState(c *telegram.Client, state LoginState) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if client == c {
		currentState = state
	}
}

// --- The Login Logic (Running in Background) ---

func startBackgroundLogin() {
	stateMu.Lock()
	c, ctx := client, clientCtx
	stateMu.Unlock()

	// Generate QR with the specific Options you requested
	qr, err := c.QRLogin(telegram.QrOptions{
		Timeout:    300,
		MaxRetries: 3,

		// 1. THIS IS CALLED IF 2FA IS ON
		PasswordCallback: func() (string, error) {
			fmt.Println("Library requested password. Waiting for Browser...")
			setClientState(c, StatePassword)

			// BLOCK HERE: Wait until browser sends password via /api/password
			select {
			case pass := <-passwordChan:
				return pass, nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		},

		// 2. THIS IS CALLED IF PASSWORD WAS WRONG
		OnWrongPassword: func(attempt, maxRetries int) bool {
			fmt.Printf("Wrong password attempt %d/%d\n", attempt, maxRetries)
			setClientState(c, StateWrongPass)
			// Return true to try again (which calls PasswordCallback again)
			return true
		},
	})

	if err != nil {
		log.Println("QR Gen Error:", err)
		setClientState(c, StateFailed)
		return
	}

	stateMu.Lock()
	if client == c {
		qrURL = qr.Url()
	}
	stateMu.Unlock()

	// Start waiting (Blocking)
	go func() {
		// This will block until scan is done AND password (if needed) is finished
		if err := qr.WaitLogin(300); err != nil {
			log.Println("Login Process Failed:", err)
			setClientState(c, StateFailed)
		} else {
			log.Println("Login Successful!")
			loadClientSession(ctx, c)
		}
	}()
}

// --- HTTP Handlers ---

func pollHandler(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()

	resp := map[string]interface{}{
		"state": currentState,
		"qr":    qrURL,
	}

	if currentState == StateLoggedIn {
		resp["user"] = currentUser
	}

	json.NewEncoder(w).Encode(resp)
}

func startLoginHandler(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	if currentState != StateIdle && currentState != StateFailed {
		stateMu.Unlock()
		json.NewEncoder(w).Encode(map[string]string{"status": "busy"})
		return
	}
	currentState = StateQR
	stateMu.Unlock()

	go startBackgroundLogin()
	time.Sleep(500 * time.Millisecond) // buffer
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func submitPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Send the password into the channel.
	// This unblocks the PasswordCallback function above.
	select {
	case passwordChan <- req.Password:
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	default:
		// Use case: User sends password but we aren't asking for it
		json.NewEncoder(w).Encode(map[string]string{"error": "Not waiting for password"})
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	c := client
	if c == nil || currentState == StateStarting {
		stateMu.Unlock()
		http.Error(w, "Telegram is still starting", http.StatusConflict)
		return
	}
	previousState := currentState
	currentState = StateStarting
	stateMu.Unlock()
	if _, err := c.AuthLogOut(); err != nil {
		setClientState(c, previousState)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	c.Disconnect()
	os.Remove(SessionFile)
	os.Remove(SessionFile + "-journal")
	initClient()
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/public/") {
			next.ServeHTTP(w, r)
			return
		}
		user, pass, ok := r.BasicAuth()

		if !ok || user != config.WebUsername || pass != config.WebPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
