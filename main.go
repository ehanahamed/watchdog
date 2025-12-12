package main

import (
    "os"
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "time"
    "strings"

    "github.com/joho/godotenv"
)

var webhookUrl string
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

func trackImage(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path

	if requirePngExt && !strings.HasSuffix(path, ".png") {
		http.NotFound(w, r)
		return
	}

    ua := r.UserAgent()
	ts := time.Now().Format("2006-01-02 15:04:05 MST")

    msg := DiscordMessage{
        Content: "someone viewed `" + path + "`\n" +
            "- time: " + ts + "\n" +
            "- user-agent: `" + ua + "`",
    }
    body, _ := json.Marshal(msg)
    http.Post(webhookUrl, "application/json", bytes.NewBuffer(body))

    w.Header().Set("Content-Type", "image/png")
    w.Write(pixel)
}

const defaultPort = "8080"
func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	requirePngExtStr := os.Getenv("REQUIRE_PNG_EXTENSION")
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
	webhookUrl = os.Getenv("WEBHOOK_URL")
	if webhookUrl == "" {
		log.Fatal(
			"WEBHOOK_URL is empty, you need to set it. \n" +
			"Check `.env` file/environment variables.",
		)
	}
    http.HandleFunc("/", trackImage)
	log.Println("Watchdog is listening on port :"+port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
