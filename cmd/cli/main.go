package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

type config struct {
	APIUrl   string
	APIToken string
	SpaceID  int
	FolderID int
}

func getConfig() (config, error) {
	var cfg config

	flag.StringVar(&cfg.APIUrl, "api-url", "https://mapi.storyblok.com/v1/", "Storyblok API URL")
	flag.StringVar(&cfg.APIToken, "api-token", "", "Storyblok API token")
	flag.IntVar(&cfg.SpaceID, "space-id", 0, "Storyblok space ID")
	flag.IntVar(&cfg.FolderID, "folder-id", 0, "Storyblok assets folder ID")

	flag.Parse()

	if cfg.APIToken == "" || cfg.SpaceID == 0 || cfg.FolderID == 0 {
		flag.Usage()
		return cfg, fmt.Errorf("api token, space id and folder id are required")
	}
	return cfg, nil
}

type AssetsResponse struct {
	Assets []Asset `json:"assets"`
}

type Asset struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
}

func getAssets(cfg config) ([]Asset, error) {
	var assets []Asset
	page := 1
	client := &http.Client{}

	for {
		assetsURL := fmt.Sprintf("%s/spaces/%d/assets?in_folder=%d&per_page=100&page=%d", cfg.APIUrl, cfg.SpaceID, cfg.FolderID, page)
		req, err := http.NewRequest(http.MethodGet, assetsURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", cfg.APIToken)
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var assetsResponse AssetsResponse
		err = json.Unmarshal(body, &assetsResponse)
		if err != nil {
			return nil, err
		}

		if len(assetsResponse.Assets) == 0 {
			break
		}

		assets = append(assets, assetsResponse.Assets...)

		page++

		if err = res.Body.Close(); err != nil {
			panic(err)
		}
	}

	return assets, nil
}

func deleteAssets(cfg config, assets []Asset) error {
	client := &http.Client{}
	bar := progressbar.Default(int64(len(assets)))
	for _, asset := range assets {
		delURL := fmt.Sprintf("%s/spaces/%d/assets/%d", cfg.APIUrl, cfg.SpaceID, asset.ID)
		req, err := http.NewRequest(http.MethodDelete, delURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", cfg.APIToken)
		res, err := client.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to delete asset %s", asset.Filename)
		}

		if err = res.Body.Close(); err != nil {
			panic(err)
		}
		if err := bar.Add(1); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := getConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get config")
	}
	allAssets, err := getAssets(cfg)
	log.Printf("You're going to delete %d assets, do you agree? Yes/No [No]", len(allAssets))

	reader := bufio.NewReader(os.Stdin)
	agreement, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read user input")
	}

	if strings.ToLower(strings.Trim(agreement, "\n")) == "yes" {
		log.Info().Msg("Deleting assets...")
		if err = deleteAssets(cfg, allAssets); err != nil {
			log.Fatal().Err(err).Msg("Failed to delete assets")
		}
		log.Info().Msg("Assets deleted")
	} else {
		log.Info().Msg("Exiting...")
		return
	}
}
