package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

type Tweet struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func main() {
	r := gin.Default()

	r.GET("/scrape", func(c *gin.Context) {
		url := c.Query("url")
		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
			return
		}

		tweets, err := scrapeTwitter(url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, tweets)
	})

	r.Run(":8080")
}

func scrapeTwitter(url string) ([]Tweet, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	// Create a timeout context
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Navigate to the URL and capture the HTML content
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second), // Wait for content to load
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}

	log.Println("HTML Content:", htmlContent) // Log the HTML content

	// Parse the HTML content to extract tweets
	tweets := parseTweets(htmlContent)
	return tweets, nil
}

func parseTweets(html string) []Tweet {
	var tweets []Tweet
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	// Use CSS selector to find the tweet elements and extract IDs and content
	doc.Find("div.css-175oi2r > div[data-testid='tweetText']").Each(func(i int, s *goquery.Selection) {
		tweetContent := s.Find("span.css-1jxf684.r-bcqeeo.r-1ttztb7.r-qvutc0.r-poiln3").Text()
		tweetLink, exists := s.Closest("article").Find("a[href^='/WuBlockchain/status/']").Attr("href")
		if exists {
			re := regexp.MustCompile(`/WuBlockchain/status/(\d+)`)
			match := re.FindStringSubmatch(tweetLink)
			if len(match) > 1 {
				tweetID := match[1]
				tweets = append(tweets, Tweet{ID: tweetID, Content: tweetContent})
			}
		}
	})

	return tweets
}
