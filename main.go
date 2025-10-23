package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/smtp"

)

type Quote struct {
    Q string `json:"q"`
    A string `json:"a"` 
}

// Fetch a random quote
func fetchQuote() (string, error) {
    resp, err := http.Get("https://zenquotes.io/api/random")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var q []Quote
    if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
        return "", err
    }
    return fmt.Sprintf("%s — %s", q[0].Q, q[0].A), nil
}

// Fetch a random image
func fetchImage() ([]byte, error) {
    resp, err := http.Get("https://picsum.photos/400")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}

// Send email with inline image
func sendEmail(subject, quote string, image []byte) error {
    from := "reynaldodomenico@gmail.com"
    password := "plba lbqu hodt sebq"
    to := "reynaldodomenico@yahoo.com"

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
    sendDailyEmail()
}
