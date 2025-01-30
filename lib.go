package vcsretriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/yuin/goldmark"
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

func GetRepos(username string) ([]Repo, error) {
	// URL pour obtenir les repos publics d'un utilisateur GitHub
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?type=pu(blic", username)

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
	return repos, nil
}

func (r Repo) String() string {
	return fmt.Sprintf(`Nom du projet : %s
Description : %s
URL du dépôt : %s
`, r.Name, r.Description, r.HTMLURL)
}

func retrievingReadme(repo Repo) ([]byte, error) {
	readmeURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/heads/master/README.md", repo.Owner.Login, repo.Name)
	readmeResp, err := http.Get(readmeURL)
	if err != nil {
		return nil, fmt.Errorf("when http.Get: %w", err)
		// fmt.Printf("Impossible de récupérer le README pour %s: %v\n", repo.Name, err)
		// continue
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

func convertRepo(repo Repo) ([]byte, error) {

	readmeBody, err := retrievingReadme(repo)
	if err != nil {
		return nil, fmt.Errorf("when retrievingReadme: %w", err)
	}
	var buf bytes.Buffer
	err = goldmark.Convert(readmeBody, &buf)
	if err != nil {
		return nil, fmt.Errorf("when goldmark.Convert: %w", err)
	}
	return buf.Bytes(), nil
	// fmt.Printf("README pour %s :\n%s\n", repo.Name, string(buf.Bytes()))
}

// Fonction pour décrire un repository GitHub à l'aide d'un modèle LLM
func DescribeGitHubRepo(ctx context.Context, model llms.Model, repo Repo, filter FilterReponse) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	readme, err := convertRepo(repo)
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
