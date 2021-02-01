package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/leveldbcache"
	"golang.org/x/oauth2"
)

func main() {
	var (
		explicitAccessToken string
		accessTokenPath     string
		username            string
		requiredOrgName     string
		credentialCacheDir  string
	)

	flag.StringVar(&explicitAccessToken, "a", "", "GitHub access token")
	flag.StringVar(&accessTokenPath, "apath", "", "GitHub access token path")
	flag.StringVar(&username, "u", "", "GitHub username")
	flag.StringVar(&requiredOrgName, "o", "", "GitHub organization to require membership in")
	flag.StringVar(&credentialCacheDir, "cpath", "", "Credential cache directory")

	flag.Parse()

	switch {
	case len(username) == 0:
		log.Fatal("username is required")

	case len(requiredOrgName) == 0:
		log.Fatal("org name is required")
	}

	ctx := context.Background()

	authedTransport := &oauth2.Transport{
		Source: oauth2.StaticTokenSource(getAccessToken(accessTokenPath, explicitAccessToken)),
	}

	client := github.NewClient(&http.Client{
		Transport: maybeCached(authedTransport, credentialCacheDir),
	})

	user_orgs, _, err := client.Organizations.List(ctx, username, nil)
	if err != nil {
		log.Fatal(err)
	}

	isInRequiredOrg := false
	for _, org := range user_orgs {
		orgName := *(org.Login)

		if orgName == requiredOrgName {
			isInRequiredOrg = true
			break
		}
	}

	if !isInRequiredOrg {
		os.Exit(0)
	}

	keys, _, err := client.Users.ListKeys(ctx, username, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, key := range keys {
		material := *(key.Key)
		fmt.Printf("%v\n", material)
	}
}

func getAccessToken(tokenPath string, explicitToken string) *oauth2.Token {
	if len(tokenPath) == 0 {
		if len(explicitToken) == 0 {
			log.Fatal("access token is required")
		}

		return &oauth2.Token{AccessToken: explicitToken}
	}

	if len(explicitToken) > 0 {
		log.Fatal("cannot set both -a and -apath")
	}

	tokenBuf, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		log.Fatal(err)
	}

	tokenStr := strings.TrimSuffix(string(tokenBuf), "\n")
	return &oauth2.Token{AccessToken: tokenStr}
}

func maybeCached(backingTransport http.RoundTripper, cacheDir string) http.RoundTripper {
	if len(cacheDir) == 0 {
		return backingTransport
	}

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return backingTransport
	}

	ldbCache, err := leveldbcache.New(path.Join(cacheDir, "credentials.db"))
	if err != nil {
		return backingTransport
	}

	return &httpcache.Transport{
		Transport: backingTransport,
		Cache:     ldbCache,
	}
}
