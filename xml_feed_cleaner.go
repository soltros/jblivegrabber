package main

import (
    "encoding/xml"
    "fmt"
    "html"
    "io/ioutil"
    "os"
    "strings"
)

// Define the structure of the podcast XML feed, matching your existing structure.
type Rss struct {
    XMLName     xml.Name `xml:"rss"`
    Version     string   `xml:"version,attr"`
    ITunesNS    string   `xml:"xmlns:itunes,attr"`
    Channel     Channel  `xml:"channel"`
}

type Channel struct {
    Title          string         `xml:"title"`
    Link           string         `xml:"link"`
    Description    string         `xml:"description"`
    Language       string         `xml:"language"`
    ITunesAuthor   string         `xml:"itunes:author"`
    ITunesExplicit string         `xml:"itunes:explicit"`
    ITunesImage    ITunesImage    `xml:"itunes:image"`
    ITunesCategory ITunesCategory `xml:"itunes:category"`
    Items          []Item         `xml:"item"`
}

type ITunesImage struct {
    Href string `xml:"href,attr"`
}

type ITunesCategory struct {
    Text string `xml:"text,attr"`
}

type Item struct {
    Title          string    `xml:"title"`
    Link           string    `xml:"link"`
    Description    string    `xml:"description"`
    PubDate        string    `xml:"pubDate"`
    Enclosure      Enclosure `xml:"enclosure"`
    ITunesExplicit string    `xml:"itunes:explicit"`
}

type Enclosure struct {
    URL    string `xml:"url,attr"`
    Length int    `xml:"length,attr"`
    Type   string `xml:"type,attr"`
}

// cleanText cleans up text by replacing underscores, decoding HTML entities,
// and replacing common HTML entities not handled by html.UnescapeString.
func cleanText(text string) string {
    // Replace underscores with spaces
    text = strings.ReplaceAll(text, "_", " ")

    // Decode HTML entities
    text = html.UnescapeString(text)

    // Manually replace common HTML entities
    replacements := map[string]string{
        "&#39;": "'",
        // Add more replacements here if needed
    }

    for old, new := range replacements {
        text = strings.ReplaceAll(text, old, new)
    }

    return text
}

func main() {
    // Read the existing XML file
    xmlFile, err := os.Open("podcast_feed.xml")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer xmlFile.Close()

    bytes, _ := ioutil.ReadAll(xmlFile)

    var rss Rss
    xml.Unmarshal(bytes, &rss)

    // Clean and update titles and descriptions
    for i, item := range rss.Channel.Items {
        rss.Channel.Items[i].Title = cleanText(item.Title)
        rss.Channel.Items[i].Description = cleanText(item.Description)
    }

    // Marshal back to XML
    output, err := xml.MarshalIndent(rss, "", "    ")
    if err != nil {
        fmt.Println("Error marshalling to XML:", err)
        return
    }

    // Write the updated XML back to the file
    err = ioutil.WriteFile("podcast_feed.xml", output, 0644)
    if err != nil {
        fmt.Println("Error writing to file:", err)
        return
    }

    fmt.Println("podcast_feed.xml has been updated.")
}
