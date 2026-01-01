package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/config"
	"github.com/darkdeathoriginal/gogrambot/handler"
	_ "github.com/darkdeathoriginal/gogrambot/plugins"
	"github.com/joho/godotenv"
)

// --- Global State ---
type LoginState string

const (
	StateIdle      LoginState = "IDLE"
	StateQR        LoginState = "QR_SCAN"
	StatePassword  LoginState = "PASSWORD_REQUIRED" // 2FA Needed
	StateWrongPass LoginState = "PASSWORD_WRONG"    // 2FA Wrong
	StateLoggedIn  LoginState = "LOGGED_IN"
	StateFailed    LoginState = "FAILED"
)

var (
	client       *telegram.Client
	currentState = StateIdle
	qrURL        = ""
	passwordChan = make(chan string) // The bridge between Browser and Callback
	stateMu      sync.Mutex
)

const SessionFile = "telegram.session"

func main() {
	godotenv.Load()

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
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func initClient() {
	appIDStr := config.AppID
	appHash := config.AppHash
	appID, _ := strconv.Atoi(appIDStr)

	var err error
	client, err = telegram.NewClient(telegram.ClientConfig{
		AppID:   int32(appID),
		AppHash: appHash,
		Session: SessionFile, // Saves login to file
	})

	if err != nil {
		log.Fatal("Client init failed:", err)
	}

	if err := client.Connect(); err != nil {
		log.Println("Connect error:", err)
	}

	// Check if session is already valid
	if me, err := client.GetMe(); err == nil {
		updateState(StateLoggedIn)
		client.SendMessage("me", fmt.Sprintf("Hello, %s!", me.FirstName))
		client.SetCommandPrefixes(config.CommandPrefix)
		selfFilter := telegram.Any(telegram.FilterOutgoing,telegram.FromUser(client.Me().ID))
		for _, plugin := range handler.Plugins {
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

			client.On(event, plugin.Handler,finalFilter)
		}
		client.On("cmd:anything", func(m *telegram.NewMessage) error {
			m.Reply("You said: " + m.Text())
			return nil
		}, telegram.FilterOutgoing)

	} else {
		updateState(StateIdle)
	}
}

// --- The Login Logic (Running in Background) ---

func startBackgroundLogin() {
	updateState(StateQR)

	// Generate QR with the specific Options you requested
	qr, err := client.QRLogin(telegram.QrOptions{
		Timeout:    300,
		MaxRetries: 3,

		// 1. THIS IS CALLED IF 2FA IS ON
		PasswordCallback: func() (string, error) {
			fmt.Println("Library requested password. Waiting for Browser...")
			updateState(StatePassword)

			// BLOCK HERE: Wait until browser sends password via /api/password
			pass := <-passwordChan

			fmt.Println("Password received from browser, returning to library...")
			return pass, nil
		},

		// 2. THIS IS CALLED IF PASSWORD WAS WRONG
		OnWrongPassword: func(attempt, maxRetries int) bool {
			fmt.Printf("Wrong password attempt %d/%d\n", attempt, maxRetries)
			updateState(StateWrongPass) // Tell browser to show "Wrong Password" error
			// Return true to try again (which calls PasswordCallback again)
			return true
		},
	})

	if err != nil {
		log.Println("QR Gen Error:", err)
		updateState(StateFailed)
		return
	}

	qrURL = qr.Url()

	// Start waiting (Blocking)
	go func() {
		// This will block until scan is done AND password (if needed) is finished
		if err := qr.WaitLogin(300); err != nil {
			log.Println("Login Process Failed:", err)
			updateState(StateFailed)
		} else {
			log.Println("Login Successful!")
			updateState(StateLoggedIn)
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
		me, _ := client.GetMe()
		resp["user"] = me
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
	client.AuthLogOut()
	os.Remove(SessionFile)
	os.Remove(SessionFile + "-journal")
	initClient()
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func updateState(s LoginState) {
	stateMu.Lock()
	currentState = s
	stateMu.Unlock()
}
