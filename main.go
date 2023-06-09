package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	miniflux "miniflux.app/client"
)

type Entry struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type EntriesResponse struct {
	Entries []Entry `json:"entries"`
}

func main() {
	minifluxURL := os.Getenv("MINIFLUX_URL")
	token := os.Getenv("MINIFLUX_TOKEN")
	receiverEmail := os.Getenv("RECEIVER_EMAIL")
	gmailEmail := os.Getenv("GMAIL_EMAIL")
	gmailPassword := os.Getenv("GMAIL_PASSWORD")
	selected_category := os.Getenv("CATEGORY")

	client := miniflux.New(minifluxURL, token)

	categories, err := client.Categories()
	if err != nil {
		fmt.Println(err)
	}

	var category_id int64

	for _, category := range categories {
		if category.Title == selected_category {
			category_id = category.ID
		}
	}

	fmt.Printf("Detected category_id is %d", category_id)

	entries, err := client.CategoryEntries(category_id, &miniflux.Filter{Status: miniflux.EntryStatusUnread, CategoryID: category_id})
	if err != nil {
		fmt.Println(err)
	}

	if entries.Total == 0 {
		log.Println("No unread entries found")
		return
	}

	emailBody := formatEmailBody(entries)
	err = sendEmail(gmailEmail, gmailPassword, receiverEmail, emailBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.MarkCategoryAsRead(category_id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Category marked as read")
	}
}

func formatEmailBody(entries *miniflux.EntryResultSet) string {
	var buffer bytes.Buffer

	for _, entry := range entries.Entries {
		buffer.WriteString(fmt.Sprintf("<h2><a href=\"%s\">%s</a></h2><br/>", entry.URL, entry.Title))
		buffer.WriteString(fmt.Sprintf("<div>%s</div>", entry.Content))
		buffer.WriteString("<hr>")
	}

	return buffer.String()
}

func sendEmail(gmailEmail, gmailPassword, toEmail, body string) error {
	auth := smtp.PlainAuth("", gmailEmail, gmailPassword, "smtp.gmail.com")

	currentDate := time.Now().Format("2006-01-02")
	subject := fmt.Sprintf("📰 News Updates - %s", currentDate)

	fmt.Println("sending email to: ", toEmail)

	to := []string{toEmail}
	msg := []byte("To: <" + toEmail + ">\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8" + "\r\n" +
		"\r\n" +
		body)

	return smtp.SendMail("smtp.gmail.com:587", auth, gmailEmail, to, msg)
}
