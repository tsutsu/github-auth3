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

	"github.com/google/go-github/v28/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/leveldbcache"
	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
)

func getDBPath() string {
	if os.Geteuid() == 0 {
		return "/var/cache/github-auth3/db"
	} else {
		config_dirs := configdir.New("tsutsu", "github-auth3")
		cache_dir := config_dirs.QueryCacheFolder()
		return path.Join(cache_dir.Path, "db")
	}
}

func main() {
	var access_token string
	var access_token_path string
	var username string
	var required_org_name string

	flag.StringVar(&access_token, "a", "", "Github access token")
	flag.StringVar(&access_token_path, "apath", "", "Github access token path")
	flag.StringVar(&required_org_name, "o", "", "Github organization to require membership in")
	flag.StringVar(&username, "u", "", "Github username")

	flag.Parse()

	if len(access_token_path) > 0 {
		if len(access_token) > 0 {
			log.Fatal("cannot set both -a and -apath")
		}

		access_token_buf, err := ioutil.ReadFile(access_token_path)
		if err != nil {
			log.Fatal(err)
		}

		access_token = strings.TrimSuffix(string(access_token_buf), "\n")
	}

	switch {
	case len(username) == 0:
		log.Fatal("username is required")

	case len(required_org_name) == 0:
		log.Fatal("org name is required")

	case len(access_token) == 0:
		log.Fatal("access token is required")
	}

	token := &oauth2.Token{AccessToken: access_token}
	auth_transport := &oauth2.Transport{Source: oauth2.StaticTokenSource(token)}

	db_path := getDBPath()
	if err := os.MkdirAll(db_path, 0700); err != nil {
		log.Fatal(err)
	}

	db_cache, err := leveldbcache.New(db_path)
	if err != nil {
		log.Fatal(err)
	}

	caching_transport := &httpcache.Transport{
		Transport: auth_transport,
		Cache:     db_cache,
	}

	client := github.NewClient(&http.Client{
		Transport: caching_transport,
	})

	ctx := context.Background()

	user_orgs, _, err := client.Organizations.List(ctx, username, nil)
	if err != nil {
		log.Fatal(err)
	}

	in_required_org := false
	for _, org := range user_orgs {
		org_name := *(org.Login)

		if org_name == required_org_name {
			in_required_org = true
			break
		}
	}

	if !in_required_org {
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
