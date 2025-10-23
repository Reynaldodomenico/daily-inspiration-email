package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/smtp"
    "os"
    "time"

    "github.com/robfig/cron/v3"
    "github.com/joho/godotenv"
)

type Quote struct {
    Content string `json:"content"`
    Author  string `json:"author"`
}

// Unsplash API response struct
type UnsplashImage struct {
    URLs struct {
        Regular string `json:"regular"`
    } `json:"urls"`
}

// Fetch a random quote using ZenQuotes (safer than Quotable)
func fetchQuote() (string, error) {
    resp, err := http.Get("https://zenquotes.io/api/random")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var q []struct {
        Q string `json:"q"`
        A string `json:"a"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
        return "", err
    }

    return fmt.Sprintf("%s — %s", q[0].Q, q[0].A), nil
}

// Fetch a random scenery image from Unsplash
func fetchImage() ([]byte, error) {
    apiKey := os.Getenv("UNSPLASH_API_KEY")
    resp, err := http.Get(fmt.Sprintf("https://api.unsplash.com/photos/random?query=scenery&client_id=%s", apiKey))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var img UnsplashImage
    if err := json.NewDecoder(resp.Body).Decode(&img); err != nil {
        return nil, err
    }

    // Fetch the actual image bytes
    imgResp, err := http.Get(img.URLs.Regular)
    if err != nil {
        return nil, err
    }
    defer imgResp.Body.Close()

    return io.ReadAll(imgResp.Body)
}

// Send email with inline image
func sendEmail(subject, quote string, image []byte) error {
    from := os.Getenv("FROM_EMAIL")
    password := os.Getenv("FROM_EMAIL_PASSWORD")
    to := os.Getenv("TO_EMAIL")

    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    imgBase64 := base64.StdEncoding.EncodeToString(image)

    body := fmt.Sprintf(`
--boundary123
Content-Type: text/html; charset="UTF-8"

<html>
<body>
<h2>Your Daily Quote</h2>
<p>%s</p>
<img src="cid:image1">
</body>
</html>

--boundary123
Content-Type: image/jpeg
Content-Transfer-Encoding: base64
Content-ID: <image1>

%s
--boundary123--`, quote, imgBase64)

    msg := []byte("From: " + from + "\r\n" +
        "To: " + to + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "MIME-Version: 1.0\r\n" +
        "Content-Type: multipart/related; boundary=boundary123\r\n\r\n" +
        body)

    auth := smtp.PlainAuth("", from, password, smtpHost)
    return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
}

// Job to send daily email
func sendDailyEmail() {
    quote, err := fetchQuote()
    if err != nil {
        log.Println("Error fetching quote:", err)
        return
    }

    image, err := fetchImage()
    if err != nil {
        log.Println("Error fetching image:", err)
        return
    }

    if err := sendEmail("Your Daily Inspiration", quote, image); err != nil {
        log.Println("Error sending email:", err)
        return
    }

    log.Println("Daily email sent successfully!")
}

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }

    // Set timezone GMT+2
    loc, err := time.LoadLocation("Etc/GMT-2")
    if err != nil {
        log.Fatal("Error loading timezone:", err)
    }

    c := cron.New(cron.WithLocation(loc))

    // Schedule job at 8:00 AM GMT+2 daily (europe/berlin time)
    _, err = c.AddFunc("0 8 * * *", sendDailyEmail)
    if err != nil {
        log.Fatal("Error scheduling cron job:", err)
    }

    c.Start()

    log.Println("Scheduler running. Waiting for 8:00 AM Berlin time to send email")

    select {}
}

