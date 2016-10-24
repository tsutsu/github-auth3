package main

import (
	"os"
	"log"
	"path"
	"net/http"
	"fmt"
	"flag"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/leveldbcache"
	"golang.org/x/oauth2"
	"github.com/shibukawa/configdir"
)

func getDBPath() string {
	if os.Geteuid() == 0 {
		return "/var/cache/github-auth3/db"
	} else {
		config_dirs := configdir.New("tsutsu", "github-auth3")
		cache_dir   := config_dirs.QueryCacheFolder()
		return path.Join(cache_dir.Path, "db")
	}
}

func main() {
	var access_token string
	var username string
	var required_org_name string

	flag.StringVar(&access_token, "a", "", "Github access token")
	flag.StringVar(&required_org_name, "o", "", "Github organization to require membership in")
	flag.StringVar(&username, "u", "", "Github username")

	flag.Parse()

	if len(username) == 0 {
		log.Fatal("username is required")
	}

	if len(required_org_name) == 0 {
		log.Fatal("org name is required")
	}

	if len(access_token) == 0 {
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

	// hoarding_transport := &apiproxy.RevalidationTransport{
	// 	Transport: caching_transport,
	// 	Check: (&githubproxy.MaxAge{
	// 		User:         time.Hour * 24,
	// 		Repository:   time.Hour * 24,
	// 		Repositories: time.Hour * 24,
	// 		Activity:     time.Hour * 12,
	// 	}).Validator(),
	// }

	client := github.NewClient(&http.Client{
		Transport: caching_transport,
	})



	user_orgs, _, err := client.Organizations.List("tsutsu", nil)
	if err != nil {
		log.Fatal(err)
	}

	user_orgs_set := make(map[string]bool)
	for _, org := range user_orgs {
		org_name := *(org.Login)
		user_orgs_set[org_name] = true
	}

	if !user_orgs_set[required_org_name] {
		os.Exit(0)
	}


	keys, _, err := client.Users.ListKeys(username, nil)
	if err != nil {
		log.Fatal(err)
	}

	var keys_materials []string
	for _, key := range keys {
		material := *(key.Key)
		keys_materials = append(keys_materials, material)
	}

	for _, material := range keys_materials {
		fmt.Printf("%v\n", material)
	}
}
