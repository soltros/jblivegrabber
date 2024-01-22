package main

import (
    "encoding/xml"
    "io/ioutil"
    "net/url"
    "os"
    "path/filepath"
    "time"
)

// Define the RSS feed with iTunes specific tags
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

func generatePodcastXML(podcastsDir string, itemDescriptions map[string]string) error {
    files, err := ioutil.ReadDir(podcastsDir)
    if err != nil {
        return err
    }

    var items []Item
    for _, f := range files {
        if filepath.Ext(f.Name()) == ".mp3" {
            title := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
            description, ok := itemDescriptions[title]
            if !ok {
                description = "No description available"
            }

            enclosureURL := "http://ubuntu-server/jblivegrabber/podcasts/" + url.PathEscape(f.Name()) // Replace with actual URL
            item := Item{
                Title:          title,
                Description:    description,
                Link:           enclosureURL,
                PubDate:        f.ModTime().Format("Mon, 02 Jan 2006 15:04:05 MST"),
                Enclosure:      Enclosure{URL: enclosureURL, Length: int(f.Size()), Type: "audio/mpeg"},
                ITunesExplicit: "no",
            }
            items = append(items, item)
        }
    }

    channel := Channel{
        Title:          "Jupiter Broadcasting Livestreams",
        Link:           "https://jupiter.tube", // Replace with your podcast's link
        Description:    "The latest Jupiter Broadcasting livestreams in MP3 form.",
        Language:       "en-us",
        ITunesAuthor:   "Jupiter Broadcasting",
        ITunesExplicit: "no",
        ITunesImage:    ITunesImage{Href: "https://static.feedpress.com/logo/allshows-5ca8d355b9812.png"}, // Replace with your podcast image URL
        ITunesCategory: ITunesCategory{Text: "Podcast Category"},
        Items:          items,
    }

    rss := Rss{
        Version:  "2.0",
        ITunesNS: "http://www.itunes.com/dtds/podcast-1.0.dtd",
        Channel:  channel,
    }

    outputFile, err := os.Create(filepath.Join(podcastsDir, "podcast_feed.xml"))
    if err != nil {
        return err
    }
    defer outputFile.Close()

    encoder := xml.NewEncoder(outputFile)
    encoder.Indent("", "    ")
    return encoder.Encode(rss)
}
