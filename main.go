package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	"gopkg.in/ini.v1"
)

func register() error {
	registerCmd := flag.NewFlagSet("register", flag.ExitOnError)
	server := registerCmd.String("server", "", "URL of server to register on")

	registerCmd.Parse(os.Args[2:])
	if *server == "" {
		return errors.New("Server must be specified")
	}

	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     *server,
		ClientName: "fedi-picbot",
		// the go-mastodon library hardcodes these scopes when authenticating, so we have to use the same ones
		Scopes:  "read write follow",
		Website: "https://github.com/archaeme/fedi-picbot",
	})
	if err != nil {
		return err
	}

	fmt.Println("Copy these into config.ini:")
	fmt.Printf("Server = %s\n", *server)
	fmt.Printf("ClientID = %s\n", app.ClientID)
	fmt.Printf("ClientSecret = %s\n", app.ClientSecret)
	fmt.Println("Don't forget to fill in your username and password in the config file!")

	return nil
}

func post() error {
	postCmd := flag.NewFlagSet("post", flag.ExitOnError)
	var workingDir string
	var configFile string
	var sourcesFile string
	postCmd.StringVar(&workingDir, "dir", ".", "Directory of config and sources file (Default: current dir)")
	postCmd.StringVar(&configFile, "config", "", "Path to config file (Default: $dir/config.ini)")
	postCmd.StringVar(&sourcesFile, "sources", "", "Path ro sources.txt file (Default: $dir/sources.txt)")
	postCmd.Parse(os.Args[2:])

	if configFile == "" {
		configFile = filepath.Join(workingDir, "config.ini")
	}

	if sourcesFile == "" {
		sourcesFile = filepath.Join(workingDir, "sources.txt")
	}

	config, err := ini.Load(configFile)
	if err != nil {
		return err
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
		return err
	}

	// TODO: allow this to be configured
	imagesDir := filepath.Join(workingDir, "images")
	img, err := getImage(sourcesFile, imagesDir)
	if err != nil {
		return err
	}
	defer img.reader.Close()

	fmt.Printf("Posting image from %s\n", img.url)

	attachment, err := client.UploadMediaFromReader(context.Background(), img.reader)
	if err != nil {
		return err
	}

	_, err = client.PostStatus(context.Background(), &mastodon.Toot{
		Status:    fmt.Sprintf("Source: %s", img.source),
		MediaIDs:  []mastodon.ID{attachment.ID},
		Sensitive: img.sensitive,
	})
	if err != nil {
		return err
	}

	return nil
}

type image struct {
	url       string
	reader    io.ReadCloser
	source    string
	sensitive bool
}

func getImage(sourcesFile string, imagesDir string) (*image, error) {
	file, err := os.Open(sourcesFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	var pick string
	var pickNum int
	for scanner.Scan() {
		line := scanner.Text()
		randomSrc := rand.NewSource(time.Now().UnixNano())
		random := rand.New(randomSrc)

		roll := random.Intn(lineNum)
		if roll == 0 {
			pick = line
			pickNum = lineNum
		}

		lineNum += 1
	}

	// Line format is <image url> <sensitive bool> <source>
	// each separated by tabs
	parts := strings.SplitN(pick, "\t", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("Line %d in sources.txt is not valid", pickNum)
	}

	url := parts[0]
	sensitive, err := strconv.ParseBool(parts[1])
	if err != nil {
		return nil, err
	}
	source := parts[2]

	var reader io.ReadCloser
	urlIsHttp := strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
	if urlIsHttp {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("Unable to fetch image, received status %s", resp.Status)
		}

		reader = resp.Body
	} else {
		url = fmt.Sprintf("%s/%s", imagesDir, url)
		reader, err = os.Open(url)
		if err != nil {
			return nil, err
		}
	}

	return &image{
		url:       url,
		reader:    reader,
		source:    source,
		sensitive: sensitive,
	}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Must use 'post' or 'register' subcommands")
		os.Exit(1)
	}

	var err error = nil
	switch os.Args[1] {
	case "register":
		err = register()
	case "post":
		err = post()
	default:
		err = errors.New("Must use 'post' or 'register' subcommands")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
