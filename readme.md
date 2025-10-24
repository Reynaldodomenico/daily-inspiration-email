# Daily Inspiration Email Sender

A lightweight Go application that sends a daily email containing:
- a random inspirational quote, and
- a random scenery image (from Unsplash).

The app is automated with GitHub Actions, which runs it every day at 08:00 (Berlin time)

---

**Features**
- Fetches a random quote from ZenQuotes API.
- Fetches random scenery images from Unsplash API.
- Sends inline images in the email using SMTP.
- Schedules email delivery using cron.
