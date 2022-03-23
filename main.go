package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	"gopkg.in/ini.v1"
)

func register() {
	registerCmd := flag.NewFlagSet("register", flag.ExitOnError)
	server := registerCmd.String("server", "", "URL of server to register on")

	registerCmd.Parse(os.Args[2:])
	if *server == "" {
		log.Fatalln("Server must be specified")
	}

	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     *server,
		ClientName: "fedi-picbot",
		// the go-mastodon library hardcodes these scopes when authenticating, so we have to use the same ones
		Scopes:  "read write follow",
		Website: "https://github.com/archaeme/fedi-picbot",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Copy these into config.ini:")
	fmt.Printf("Server = %s\n", *server)
	fmt.Printf("ClientID = %s\n", app.ClientID)
	fmt.Printf("ClientSecret = %s\n", app.ClientSecret)
	fmt.Println("Don't forget to fill in your username and password in the config file!")
}

func post() {
	postCmd := flag.NewFlagSet("post", flag.ExitOnError)
	configFile := postCmd.String("config", "config.ini", "Path to config file. Defaults to config.ini in working dir")
	sourcesFile := postCmd.String("sources", "sources.txt", "Text file that contains files to download and a link to the original source")
	postCmd.Parse(os.Args[2:])

	config, err := ini.Load(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	username := config.Section("Login").Key("Username").String()
	password := config.Section("Login").Key("Password").String()
	server := config.Section("").Key("Server").String()
	clientID := config.Section("").Key("ClientID").String()
	clientSecret := config.Section("").Key("ClientSecret").String()

	client := mastodon.NewClient(&mastodon.Config{
		Server:       server,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	err = client.Authenticate(context.Background(), username, password)
	if err != nil {
		log.Fatal(err)
	}

	img := getImage(*sourcesFile)
	log.Printf("Posting image from %s\n", img.URL)
	resp, err := http.Get(img.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	attachment, err := client.UploadMediaFromReader(context.Background(), resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.PostStatus(context.Background(), &mastodon.Toot{
		Status:    fmt.Sprintf("Source: %s", img.Source),
		MediaIDs:  []mastodon.ID{attachment.ID},
		Sensitive: img.Sensitive,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Image struct {
	URL       string
	Source    string
	Sensitive bool
}

func getImage(sourcesFile string) *Image {
	file, err := os.Open(sourcesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	var pick string
	for scanner.Scan() {
		line := scanner.Text()
		randomSrc := rand.NewSource(time.Now().UnixNano())
		random := rand.New(randomSrc)

		roll := random.Intn(lineNum)
		if roll == 0 {
			pick = line
		}

		lineNum += 1
	}

	// Line format is <image url> <source url> <sensitive bool>
	// each separated by tabs
	// FIXME: maybe this could be a sqlite db or someting that isn't error prone
	parts := strings.SplitN(pick, "\t", 3)
	url := parts[0]
	source := parts[1]
	sensitive, err := strconv.ParseBool(parts[2])
	if err != nil {
		log.Fatal(err)
	}
	return &Image{
		URL:       url,
		Source:    source,
		Sensitive: sensitive,
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Must use 'post' or 'register' subcommands")
	}

	switch os.Args[1] {
	case "register":
		register()
	case "post":
		post()
	default:
		log.Fatalln("Must use 'post' or 'register' subcommands")
	}
}
