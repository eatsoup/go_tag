/*
Simple remote registry tagger to speed up your CI/CD
BIG TODO: Fix authentication
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var authenticated bool
var registryUser string
var registryPassword string

type authToken struct {
	Token string `json:"token"`
}

func main() {
	var authToken string
	var oldTag string
	var newTag string
	var registry string
	if len(os.Args) < 3 {
		fmt.Println("This application requires at least 2 parameters")
		fmt.Printf("Example: %s register.example.com/nginx:dev register.example.com/nginx:1.0\n", os.Args[0])
		os.Exit(1)
	}
	oldTag = os.Args[1]
	newTag = os.Args[2]
	registry = os.Getenv("REGISTRY")
	if os.Getenv("REGISTRY_USER") == "" && os.Getenv("REGISTRY_PASSWORD") == "" {
		fmt.Println("REGISTRY_USER and REGISTRY_PASSWORD environment variable not found, assuming unauthenticated")
		authenticated = false
	} else {
		fmt.Println("REGISTRY_USER or REGISTRY_PASSWORD environment variable found, using authentication")
		authenticated = true
		registryUser = os.Getenv("REGISTRY_USER")
		registryPassword = os.Getenv("REGISTRY_PASSWORD")
		authToken = getAuthToken(registry, registryUser, registryPassword, oldTag)
	}
	image, body := fetchManifest(registry, authToken, oldTag)
	setTag(registry, authToken, image, newTag, body)

}

func fetchManifest(registry, authToken, tag string) (string, []byte) {
	if registry == "" {
		registry = "https://registry-1.docker.io"
	}
	s := strings.Split(tag, ":")
	var image string
	var version string
	if len(s) > 1 {
		image = s[0]
		version = s[len(s)-1]
	} else {
		image = tag
		version = "latest"
	}
	url := registry + "/v2/" + image + "/manifests/" + version
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	if authToken != "" {
		request.Header.Set("Authorization", "Bearer "+authToken)
	}
	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	} else {
		if r.StatusCode != 200 && r.StatusCode != 201 {
			fmt.Println("Error fetching manifest")
			fmt.Println(r.StatusCode)
		} else {
			body, _ := ioutil.ReadAll(r.Body)
			return image, body
		}
	}
	return image, nil
}

func setTag(registry, authToken, image, newVersion string, manifest []byte) {
	var version string
	var response string
	if registry == "" {
		registry = "https://registry-1.docker.io"
	}
	s := strings.Split(newVersion, ":")
	if len(s) > 1 {
		image = s[0]
		version = s[len(s)-1]
	} else {
		version = "latest"
	}
	fmt.Printf("Tagging %s with version: %s\n", image, version)
	url := registry + "/v2/" + image + "/manifests/" + version
	request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(manifest))
	request.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	if authToken != "" {
		request.Header.Set("Authorization", "Bearer "+authToken)
	}
	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		fmt.Println(r.StatusCode)
	} else {
		if r.StatusCode == 200 || r.StatusCode == 201 {
			fmt.Println("Remote tag complete")
		} else {
			body, _ := ioutil.ReadAll(r.Body)
			response = string(body)
			fmt.Println("Errorcode:", r.StatusCode)
			fmt.Println("Message:", response)
			os.Exit(1)
		}
	}
}

func getAuthToken(registry, username, password, tag string) string {
	var token authToken
	var readToken string
	image := strings.Split(tag, ":")[0]
	client := &http.Client{}
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	if registry == "" {
		registry = "https://auth.docker.io/token?scope=repository:" + image + ":pull,push&service=registry.docker.io"
		request, _ := http.NewRequest("GET", registry, nil)
		request.Header.Set("Content-type", "application/json")
		request.Header.Set("Authorization", "Basic "+credentials)
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal([]byte(string(body)), &token)
		readToken = token.Token
	}
	return readToken
}
