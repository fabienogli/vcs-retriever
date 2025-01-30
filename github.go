package vcsretriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// Structure pour représenter un dépôt GitHub
type Repo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	Readme []byte `json:"readme"`
}

type FilterReponse func(string) string

func GetRepos(username string) ([]Repository, error) {
	// URL pour obtenir les repos publics d'un utilisateur GitHub
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?type=public", username)

	// Requête HTTP pour obtenir les repos
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("when http.Get(%s): %w", url, err)
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("when io.ReadAll: %w", err)
	}

	// Décoder la réponse JSON
	var repos []Repo
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("when json.Unmarshal: %w", err)
	}
	repositories := make([]Repository, len(repos))
	for i, repo := range repos {
		repositories[i] = Repository{
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			URL:         repo.HTMLURL,
			User:        username,
		}
	}
	return repositories, nil
}

func (r Repo) String() string {
	return fmt.Sprintf(`Nom du projet : %s
Description : %s
URL du dépôt : %s
`, r.Name, r.Description, r.HTMLURL)
}

func RetrievingReadme(repo Repository) ([]byte, error) {
	readmeURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/heads/master/README.md", repo.User, repo.Name)
	readmeResp, err := http.Get(readmeURL)
	log.Println(readmeURL)
	if err != nil {
		return nil, fmt.Errorf("when http.Get: %w", err)
	}
	if readmeResp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found: %s", readmeURL)
	}
	defer readmeResp.Body.Close()

	readmeBody, err := io.ReadAll(readmeResp.Body)
	if err != nil {
		return nil, fmt.Errorf("when io.ReadAll: %w", err)
	}
	if len(readmeBody) > 0 {
		return readmeBody, nil
	}

	return nil, fmt.Errorf("readme not found")
}

// Fonction pour décrire un repository GitHub à l'aide d'un modèle LLM
func DescribeGitHubRepoWithIA(ctx context.Context, model llms.Model, repo Repository, filter FilterReponse) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	readme, err := ReadmeToByte(repo)
	if err != nil {
		return "", fmt.Errorf("when converting repo: %w", err)
	}
	repo.Readme = readme

	prompt := generatePrompt(repo)

	// Interroger l'LLM via Ollama
	description, err := queryOllamaLLM(ctx, model, prompt)
	if err != nil {
		return "", fmt.Errorf("when queryOllamaLLM: %w", err)
	}
	return filter(description), nil
}

// GitHub GraphQL API URL
const githubGraphQLAPI = "https://api.github.com/graphql"

// GraphQL query to get pinned repositories
const query = `{
  "query": "query { user(login: \"USERNAME\") { pinnedItems(first: 6, types: REPOSITORY) { nodes { ... on Repository { name url description stargazerCount forkCount primaryLanguage { name } } } } } }"
}`

// PinnedRepositoriesResponse represents the JSON structure from GitHub's GraphQL API
type PinnedRepositoriesResponse struct {
	Data struct {
		User struct {
			PinnedItems struct {
				Nodes []struct {
					Name            string `json:"name"`
					URL             string `json:"url"`
					Description     string `json:"description"`
					StargazerCount  int    `json:"stargazerCount"`
					ForkCount       int    `json:"forkCount"`
					PrimaryLanguage struct {
						Name string `json:"name"`
					} `json:"primaryLanguage"`
				} `json:"nodes"`
			} `json:"pinnedItems"`
		} `json:"user"`
	} `json:"data"`
}

// Replace USERNAME with the actual GitHub username
func GetPinnedRepositories(token string, username string) ([]Repository, error) {
	// Replace username dynamically
	replacedQuery := bytes.Replace([]byte(query), []byte("USERNAME"), []byte(username), -1)

	// Create a new request
	req, err := http.NewRequest("POST", githubGraphQLAPI, bytes.NewBuffer(replacedQuery))
	if err != nil {
		return nil, fmt.Errorf("when http.NewRequest: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("when client.Do: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("when io.ReadAll: %w", err)
	}
	var pinnedResponse PinnedRepositoriesResponse
	err = json.Unmarshal(body, &pinnedResponse)
	if err != nil {
		return nil, fmt.Errorf("when json.Unmarshal(%s): %w", string(body), err)
	}
	repositories := make([]Repository, len(pinnedResponse.Data.User.PinnedItems.Nodes))
	for i, node := range pinnedResponse.Data.User.PinnedItems.Nodes {
		repositories[i] = Repository{
			Name:            node.Name,
			Description:     node.Description,
			URL:             node.URL,
			User:            username,
			PrimaryLanguage: node.PrimaryLanguage.Name,
		}
	}
	return repositories, nil
}
