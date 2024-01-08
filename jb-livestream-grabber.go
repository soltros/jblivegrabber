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
)

type Rss struct {
    Channel Channel `xml:"channel"`
}

type Channel struct {
    Title       string  `xml:"title"`
    Description string  `xml:"description"`
    Link        string  `xml:"link"`
    Items       []Item  `xml:"item"`
}

type Item struct {
    Title       string    `xml:"title"`
    Link        string    `xml:"link"`
    Description string    `xml:"description"`
    PubDate     string    `xml:"pubDate"`
    Date        time.Time // Field to store parsed date
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

            item := Item{
                Title:       title,
                Description: description,
                Link:        filepath.Join(podcastsDir, f.Name()),
                PubDate:     f.ModTime().Format("Mon, 02 Jan 2006 15:04:05 MST"),
            }
            items = append(items, item)
        }
    }

    channel := Channel{
        Title:       "Jupiter Broadcasting live shows",
        Description: "Direct rips of JB livestream shows.",
        Link:        podcastsDir,
        Items:       items,
    }

    rss := Rss{Channel: channel}
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
