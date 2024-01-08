package main

import (
    "encoding/xml"
    "fmt"
    "html"
    "io/ioutil"
    "os"
    "strings"
)

// Podcast represents the structure of the podcast XML.
type Podcast struct {
    XMLName xml.Name `xml:"rss"`
    Channel Channel  `xml:"channel"`
}

// Channel represents the podcast channel information.
type Channel struct {
    Items []Item `xml:"item"`
}

// Item represents a single podcast episode.
type Item struct {
    Title       string `xml:"title"`
    Description string `xml:"description"`
    // include other fields if needed
}

// cleanText cleans up text by replacing underscores and decoding HTML entities.
func cleanText(text string) string {
    // Replace underscores with spaces
    text = strings.ReplaceAll(text, "_", " ")

    // Decode HTML entities
    text = html.UnescapeString(text)

    return text
}

func main() {
    // Read XML file
    xmlFile, err := os.Open("podcast_feed.xml")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer xmlFile.Close()

    bytes, _ := ioutil.ReadAll(xmlFile)

    var podcast Podcast
    xml.Unmarshal(bytes, &podcast)

    // Clean and update titles and descriptions
    for i, item := range podcast.Channel.Items {
        podcast.Channel.Items[i].Title = cleanText(item.Title)
        podcast.Channel.Items[i].Description = cleanText(item.Description)
    }

    // Marshal back to XML
    output, err := xml.MarshalIndent(podcast, "", "  ")
    if err != nil {
        fmt.Println("Error marshalling to XML:", err)
        return
    }

    // Output the updated XML
    fmt.Println(string(output))
}
