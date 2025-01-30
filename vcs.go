package vcsretriever

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fabienogli/vcs-retriever/github"
	"github.com/tmc/langchaingo/llms"
)

func FromRepo(repo github.Repo, username string) Repository {
	return Repository{
		Name:        repo.Name,
		FullName:    repo.FullName,
		Description: repo.Description,
		URL:         repo.HTMLURL,
		User:        username,
	}
}

func FromPinned(pinned github.PinnedRepositoriesResponse, username string) []Repository {
	repositories := make([]Repository, len(pinned.Data.User.PinnedItems.Nodes))
	for i, node := range pinned.Data.User.PinnedItems.Nodes {
		repositories[i] = Repository{
			Name:            node.Name,
			Description:     node.Description,
			URL:             node.URL,
			User:            username,
			PrimaryLanguage: node.PrimaryLanguage.Name,
		}
	}
	return repositories
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

func SummarizeRepos(ctx context.Context, username string, model llms.Model, filter FilterReponse) error {
	repos, err := github.GetRepos(username)
	if err != nil {
		return fmt.Errorf("erreur getRepos: %w", err)
	}

	newRepos := make([]Repository, len(repos))
	for i, repo := range repos {
		newRepos[i] = FromRepo(repo, username)
	}

	// Pour chaque projet, récupérer et afficher le README
	for _, repo := range newRepos {
		log.Printf("trying to describe %+v\n", repo)
		description, err := DescribeGitHubRepoWithIA(ctx, model, repo, filter)
		if err != nil {
			fmt.Printf("when describeGitHubRepo: %v\n", err)
			continue
		}
		fmt.Printf("Description generated: %s\n", description)
		fmt.Println("----------------------------------------")
	}
	return nil
}
