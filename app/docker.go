package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	authUrl     = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull"
	manifestUrl = "https://registry.hub.docker.com/v2/library/%s/manifests/%s"
	layerUrl    = "https://registry.hub.docker.com/v2/library/%s/blobs/%s"
)

type DockerAuth struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

type DockerManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

func Authenticate(image string) *DockerAuth {
	// Send a GET request to get the resp
	url := fmt.Sprintf(authUrl, image)
	resp, err := http.Get(url)
	must(err)
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	must(err)
	var auth DockerAuth
	json.Unmarshal(buf, &auth)
	return &auth
}

func GetManifest(auth *DockerAuth, image, version string) *DockerManifest {
	url := fmt.Sprintf(manifestUrl, image, version)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	must(err)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Authorization", "Bearer "+auth.Token)

	resp, err := client.Do(req)
	must(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	must(err)

	var manifest DockerManifest
	json.Unmarshal(body, &manifest)
	return &manifest
}

func Download(image string, dest string) {
	imageName, version, ok := strings.Cut(image, ":")
	if !ok {
		imageName = image
		version = "latest"
	}
	auth := Authenticate(imageName)
	manifest := GetManifest(auth, image, version)
	for i, layer := range manifest.Layers {
		downloadUrl := fmt.Sprintf(layerUrl, imageName, layer.Digest)
		filename := filepath.Join(dest, fmt.Sprintf("layer-%d.tar", i))
		DownloadLayer(auth, downloadUrl, filename)
		ExtractTarGz(filename, dest)
	}
}

func DownloadLayer(auth *DockerAuth, url string, outfile string) {
	out, err := os.Create(outfile)
	must(err)
	defer out.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	must(err)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Authorization", "Bearer "+auth.Token)

	resp, err := client.Do(req)
	must(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	must(err)
}

func ExtractTarGz(filename, dest string) {
	cmd := exec.Command("tar", "xzf", filename, "-C", dest)
	err := cmd.Run()
	must(err)
  	// Remove the tar file after extraction
	err = os.Remove(filename)
  must(err)
}
