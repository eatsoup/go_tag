/*
Simple remote registry tagger to speed up your CI/CD
BIG TODO: Fix authentication
*/

package main

import (
	"bytes"
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

type AuthToken struct {
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
	if os.Getenv("REGISTRY") != "" {
		registry = os.Getenv("REGISTRY")
	} else {
		fmt.Println("Registry env var not found, assuming docker hub")
		registry = "https://registry-1.docker.io"
	}
	if os.Getenv("REGISTRY_USER") == "" && os.Getenv("REGISTRY_PASSWORD") == "" {
		fmt.Println("REGISTRY_USER and REGISTRY_PASSWORD environment variable not found, assuming unauthenticated")
		authenticated = false
	} else {
		fmt.Println("REGISTRY_USER or REGISTRY_PASSWORD environment variable found, using authentication")
		authenticated = true
		registryUser = os.Getenv("REGISTRY_USER")
		registryPassword = os.Getenv("REGISTRY_PASSWORD")
		authToken = getAuthToken(registry, registryUser, registryPassword)
		fmt.Println(authToken)
	}
	image, body := fetchManifest(registry, authToken, oldTag)
	setTag(registry, authToken, image, body, newTag)

}

func fetchManifest(registry string, authToken string, tag string) (string, []byte) {
	s := strings.Split(tag, ":")
	var image string
	var version string
	if len(s) > 1 {
		image = strings.Join(s[0:1], ":")
		version = s[len(s)-1]
	} else {
		image = tag
		version = "latest"
	}
	fmt.Println("Registry:", registry)
	fmt.Println("Image:", image, version)
	fmt.Println("Authentication:", registryUser, registryPassword)
	request, _ := http.NewRequest("GET", registry+"/v2/"+image+"/manifests/"+version, nil)
	fmt.Println(registry+"/v2/"+image+"/manifests/"+version, nil)
	request.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	if authToken != "" {
		request.Header.Set("Authorization", "JWT "+authToken)
	}
	fmt.Println(authToken)
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

func setTag(registry string, authToken string, image string, manifest []byte, newVersion string) {
	fmt.Printf("Tagging %s with version: %s\n", image, newVersion)
	var response string
	request, _ := http.NewRequest("PUT", registry+"/v2/"+image+"/manifests/"+newVersion, bytes.NewBuffer(manifest))
	request.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	if authToken != "" {
		request.Header.Set("Authorization", "JWT "+authToken)
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

func getAuthToken(registry, username, password, image string) (string, string) {
	var token AuthToken
	var readToken string
	client := &http.Client{}
	requestBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if registry == "https://registry-1.docker.io" {
		registry = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + username + "/" + image + ":pull"
		request, _ := http.NewRequest("GET", registry, nil)
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal([]byte(string(body)), &token)
		readToken = token.Token
	}
	request, _ := http.NewRequest("POST", registry+"/v2/users/login/", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal([]byte(string(body)), &token)
	if err != nil {
		fmt.Println(err)
	}
	return readToken, token.Token
}
