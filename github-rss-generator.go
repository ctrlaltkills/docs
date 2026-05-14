package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GitHub API response structures
type GitHubUser struct {
	Login  string `json:"login"`
	URL    string `json:"html_url"`
	Avatar string `json:"avatar_url"`
}

type StarredRepo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"html_url"`
	Owner       GitHubUser `json:"owner"`
	StarredAt   string    `json:"starred_at"`
	Language    string    `json:"language"`
}

type FollowingUser struct {
	Login  string `json:"login"`
	Name   string `json:"name"`
	Bio    string `json:"bio"`
	URL    string `json:"html_url"`
	Avatar string `json:"avatar_url"`
}

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("❌ Error: GITHUB_TOKEN not set")
		fmt.Println("Get token from: https://github.com/settings/personal-access-tokens")
		os.Exit(1)
	}

	username := "ctrlaltkills"
	count := 50

	// Generate Starred RSS
	fmt.Println("📥 Fetching starred repositories...")
	starredFeed := generateStarredRSS(token, username, count)
	saveToFile("starred.xml", starredFeed)
	fmt.Println("✅ Saved: starred.xml")

	// Generate Following RSS
	fmt.Println("📥 Fetching following users...")
	followingFeed := generateFollowingRSS(token, username, count)
	saveToFile("following.xml", followingFeed)
	fmt.Println("✅ Saved: following.xml")

	fmt.Println("\n📌 RSS Feed URLs:")
	fmt.Println("   Starred:  " + getCurrentDir() + "/starred.xml")
	fmt.Println("   Following: " + getCurrentDir() + "/following.xml")
}

func generateStarredRSS(token, username string, count int) string {
	repos := fetchStarredRepos(token, username, count)

	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>` + username + ` - Starred Repositories</title>
    <link>https://github.com/` + username + `?tab=stars</link>
    <description>Latest ` + fmt.Sprintf("%d", count) + ` starred repositories</description>
    <lastBuildDate>` + time.Now().Format(time.RFC1123Z) + `</lastBuildDate>
`

	for _, repo := range repos {
		desc := repo.Description
		if desc == "" {
			desc = "No description"
		}
		lang := repo.Language
		if lang == "" {
			lang = "Unknown"
		}

		rss += `    <item>
      <title>` + repo.Owner.Login + `/` + repo.Name + `</title>
      <link>` + repo.URL + `</link>
      <description>` + escapeXML(desc) + ` | Language: ` + lang + `</description>
      <pubDate>` + parseDate(repo.StarredAt) + `</pubDate>
      <guid>` + repo.URL + `</guid>
    </item>
`
	}

	rss += `  </channel>
</rss>`
	return rss
}

func generateFollowingRSS(token, username string, count int) string {
	users := fetchFollowingUsers(token, username, count)

	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>` + username + ` - Following</title>
    <link>https://github.com/` + username + `?tab=following</link>
    <description>Users followed by ` + username + `</description>
    <lastBuildDate>` + time.Now().Format(time.RFC1123Z) + `</lastBuildDate>
`

	for _, user := range users {
		bio := user.Bio
		if bio == "" {
			bio = "No bio"
		}

		rss += `    <item>
      <title>` + user.Login + `</title>
      <link>` + user.URL + `</link>
      <description>` + escapeXML(bio) + `</description>
      <guid>` + user.URL + `</guid>
    </item>
`
	}

	rss += `  </channel>
</rss>`
	return rss
}

func fetchStarredRepos(token, username string, count int) []StarredRepo {
	url := fmt.Sprintf("https://api.github.com/users/%s/starred?per_page=%d&sort=updated", username, count)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var repos []StarredRepo
	json.Unmarshal(body, &repos)

	return repos
}

func fetchFollowingUsers(token, username string, count int) []FollowingUser {
	url := fmt.Sprintf("https://api.github.com/users/%s/following?per_page=%d", username, count)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var users []FollowingUser
	json.Unmarshal(body, &users)

	return users
}

func parseDate(dateStr string) string {
	if dateStr == "" {
		return time.Now().Format(time.RFC1123Z)
	}
	t, _ := time.Parse(time.RFC3339, dateStr)
	return t.Format(time.RFC1123Z)
}

func escapeXML(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '<':
			result += "&lt;"
		case '>':
			result += "&gt;"
		case '&':
			result += "&amp;"
		case '"':
			result += "&quot;"
		case '\'':
			result += "&apos;"
		default:
			result += string(c)
		}
	}
	return result
}

func saveToFile(filename, content string) {
	os.WriteFile(filename, []byte(content), 0644)
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}
