package main

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/joho/godotenv"
)

var webhookUrl string
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
    ua := r.UserAgent()
    ts := time.Now().Format(time.RFC3339)

    msg := DiscordMessage{
        Content: "/ viewed:\n" +
            "- Time: " + ts + "\n" +
            "- User-Agent: " + ua,
    }
    body, _ := json.Marshal(msg)
    http.Post(webhookURL, "application/json", bytes.NewBuffer(body))

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
	webhookUrl := os.Getenv("WEBHOOK_URL")
    http.HandleFunc("/watchdog.png", trackImage)
	log.Println("Listening on :"+port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
