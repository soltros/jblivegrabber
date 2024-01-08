package main

import (
    "bufio"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "time"
    "net/url"
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
    Date           time.Time // Field to store parsed date
    Enclosure      Enclosure `xml:"enclosure"`
    ITunesExplicit string    `xml:"itunes:explicit"`
}

type Enclosure struct {
    URL    string `xml:"url,attr"`
    Length int    `xml:"length,attr"`
    Type   string `xml:"type,attr"`
}

func parseRSS(url string) ([]Item, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    rss := Rss{}
    decoder := xml.NewDecoder(resp.Body)
    err = decoder.Decode(&rss)
    if err != nil {
        return nil, err
    }

    return rss.Channel.Items, nil
}

func readProcessedItems(filePath string) (map[string]bool, error) {
    file, err := os.Open(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return make(map[string]bool), nil
        }
        return nil, err
    }
    defer file.Close()

    processed := make(map[string]bool)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        processed[scanner.Text()] = true
    }
    return processed, scanner.Err()
}

func saveProcessedItem(filePath, item string) error {
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = file.WriteString(item + "\n")
    return err
}

func downloadAudio(link string, podcastsDir string, wg *sync.WaitGroup, done chan<- bool, desc string) {
    defer wg.Done()

    title := strings.ReplaceAll(strings.TrimSpace(strings.Split(desc, ":")[0]), " ", "_")
    outputPath := filepath.Join(podcastsDir, title+".mp3")

    cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "-o", outputPath, link)

    if err := cmd.Run(); err != nil {
        fmt.Printf("Error downloading audio for %s: %s\n", link, err)
        done <- false
    } else {
        done <- true
    }
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

            enclosureURL := "http://YOUR_WEB_SERVER/jblivegrabber/podcasts/" + url.PathEscape(f.Name()) // Replace with actual URL
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
        ITunesCategory: ITunesCategory{Text: "Technology"},
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

func parseAndSortItems(items []Item) ([]Item, error) {
    layout := "Mon, 02 Jan 2006 15:04:05 MST"
    for i, item := range items {
        parsedDate, err := time.Parse(layout, item.PubDate)
        if err != nil {
            return nil, err
        }
        items[i].Date = parsedDate
    }

    sort.Slice(items, func(i, j int) bool {
        return items[i].Date.After(items[j].Date)
    })

    return items, nil
}

func main() {
    rssURL := "https://jupiter.tube/feeds/videos.xml"
    processedFilePath := "processed_items.txt"
    podcastsDir := "podcasts"
    maxConcurrentDownloads := 5

    if err := os.MkdirAll(podcastsDir, os.ModePerm); err != nil {
        fmt.Println("Error creating podcasts directory:", err)
        return
    }

    items, err := parseRSS(rssURL)
    if err != nil {
        fmt.Println("Error parsing RSS feed:", err)
        return
    }

    sortedItems, err := parseAndSortItems(items)
    if err != nil {
        fmt.Println("Error parsing item dates:", err)
        return
    }

    var newestItems []Item
    for i, item := range sortedItems {
        if i >= 3 {
            break
        }
        newestItems = append(newestItems, item)
    }

    processedItems, err := readProcessedItems(processedFilePath)
    if err != nil {
        fmt.Println("Error reading processed items:", err)
        return
    }

    var wg sync.WaitGroup
    downloadSem := make(chan struct{}, maxConcurrentDownloads)
    done := make(chan bool)
    itemDescriptions := make(map[string]string)

    for _, item := range newestItems {
        if processedItems[item.Link] {
            fmt.Println("Already processed:", item.Link)
            continue
        }

        wg.Add(1)
        downloadSem <- struct{}{}

        go func(link, desc string) {
            defer func() { <-downloadSem }()
            downloadAudio(link, podcastsDir, &wg, done, desc)
        }(item.Link, item.Description)

        title := strings.ReplaceAll(strings.TrimSpace(strings.Split(item.Description, ":")[0]), " ", "_")
        itemDescriptions[title] = item.Description

        if err := saveProcessedItem(processedFilePath, item.Link); err != nil {
            fmt.Printf("Error saving processed item %s: %s\n", item.Link, err)
        }
    }

    go func() {
        wg.Wait()
        close(done)
    }()

    for range done {}

    if err := generatePodcastXML(podcastsDir, itemDescriptions); err != nil {
        fmt.Printf("Error generating podcast XML: %s\n", err)
    }
}
