package main

import (
    "io"
    "os"
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "time"
    "strings"

    "github.com/joho/godotenv"
)

type WebhookRoute struct {
    Prefix string
    URL    string
}

var webhookRoutes []WebhookRoute
var fallbackWebhook string
var requirePngExt bool
var pixel = []byte{
    // a valid 1x1 transparent PNG
    0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
    0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
    0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
    0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
    0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
    0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
    0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
    0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
    0x42, 0x60, 0x82,
}

type DiscordMessage struct {
    Content string `json:"content"`
}

func webhookForPath(path string) string {
    var bestMatch WebhookRoute
    matched := false

    for _, route := range webhookRoutes {
        if strings.HasPrefix(path, route.Prefix) {
            if !matched || len(route.Prefix) > len(bestMatch.Prefix) {
                bestMatch = route
                matched = true
            }
        }
    }

    if matched {
        return bestMatch.URL
    }

    return fallbackWebhook
}

func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max] + "\n…(truncated)"
}

func sendWebhook(webhookURL, content string) {
    if webhookURL == "" {
        return
    }

    msg := DiscordMessage{Content: content}
    body, _ := json.Marshal(msg)
    _, _ = http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
}

func track(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

    path := r.URL.Path
    ua := r.UserAgent()
    ts := time.Now().Format("2006-01-02 · 15:04 MST")

    webhook := webhookForPath(path)

    switch r.Method {

    case http.MethodGet:
        if requirePngExt && !strings.HasSuffix(path, ".png") {
            http.NotFound(w, r)
            return
        }

        go sendWebhook(
            webhook,
            "someone viewed `" + path + "`\n\n" +
                ts + "\n" +
                "User-Agent: `" + ua + "`",
        )

        w.Header().Set("Content-Type", "image/png")
        w.Write(pixel)

    case http.MethodPost:
        bodyBytes, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "failed to read body", http.StatusBadRequest)
            return
        }

        go sendWebhook(
            webhook,
            "message received at `" + path + "`\n" +
                ts + "\n" +
                "User-Agent: `" + ua + "`\n\n" +
				truncate(string(bodyBytes), 1800),
        )

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("okayy"))

    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

const defaultPort = "8080"
func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	requirePngExtStr := strings.ToLower(os.Getenv("REQUIRE_PNG_EXTENSION"))
	if requirePngExtStr == "true" {
		requirePngExt = true
	} else if requirePngExtStr == "false" || requirePngExtStr == "" {
		requirePngExt = false
	} else {
		log.Fatal(
			"REQUIRE_PNG_EXTENSION must be \"true\", \"false\", or empty. \n" +
			"Check `.env` file/environment variables.",
		)
	}

	routes := os.Getenv("WEBHOOK_ROUTES")
	for _, entry := range strings.Split(routes, ";") {
	    if entry == "" {
	        continue
	    }
	    parts := strings.SplitN(entry, "=", 2)
	    if len(parts) != 2 {
	        log.Fatalf("invalid WEBHOOK_ROUTES entry: %s", entry)
	    }
		prefix := strings.TrimRight(parts[0], "/")
	    webhookRoutes = append(webhookRoutes, WebhookRoute{
	        Prefix: prefix,
	        URL:    parts[1],
	    })
	}
	
	fallbackWebhook = os.Getenv("FALLBACK_WEBHOOK")
	if len(webhookRoutes) == 0 && fallbackWebhook == "" {
	    log.Fatal("No WEBHOOK_ROUTES or FALLBACK_WEBHOOK configured")
	}

    http.HandleFunc("/", track)
	log.Println("Watchdog is listening on port :"+port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
