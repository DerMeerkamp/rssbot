package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Define the structure of the RSS feed
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// Config structure to hold RSS feed URLs and checking interval
type Config struct {
	Feeds           []FeedConfig `yaml:"feeds"`
	IntervalMinutes int          `yaml:"interval_minutes"`
}

type FeedConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

func main() {
	// Load configuration
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Info().Msg("Start Rssbot")
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config:")
		return
	} else {
		log.Info().Msgf("Current config: %s", config)
	}

	// Map to store last GUIDs for each feed
	lastGUIDs := make(map[string]map[string]bool)

	for {
		for _, feedConfig := range config.Feeds {
			// Check for new entries in each feed
			fmt.Println("fetch feed:", feedConfig.URL)
			rss, err := fetchRSS(feedConfig.URL)
			if err != nil {
				fmt.Printf("Error fetching RSS feed (%s): %s\n", feedConfig.Name, err)
				continue
			}

			// Initialize the GUID map for this feed if it doesn't exist
			if lastGUIDs[feedConfig.URL] == nil {

				lastGUIDs[feedConfig.URL] = make(map[string]bool)
				for _, item := range rss.Channel.Items {
					lastGUIDs[feedConfig.URL][item.GUID] = true
				}
			} else {
				for _, item := range rss.Channel.Items {
					if _, exists := lastGUIDs[feedConfig.URL][item.GUID]; !exists {
						// Process the new item
						fmt.Printf("New Item Found in %s: %s (%s)\n", feedConfig.Name, item.Title, item.Link)
						lastGUIDs[feedConfig.URL][item.GUID] = true
					}
				}
			}
		}

		// Wait for the specified interval before checking again
		time.Sleep(time.Duration(config.IntervalMinutes) * time.Minute)
	}
}

// Function to load configuration from a YAML file
func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Function to fetch and parse the RSS feed
func fetchRSS(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return nil, err
	}

	return &rss, nil
}
