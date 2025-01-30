package main

import (
	"context"
	"fmt"
	"log"

	vcsretriever "github.com/fabienogli/vcs-retriever"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	ctx := context.Background()
	//specific to deepseek
	filter, err := vcsretriever.DeepseekFilter()
	if err != nil {
		log.Fatalf("when deepseekFilter: %v", err)
		return
	}
	model := "deepseek-r1:1.5b"

	//github username
	username := "fabienogli"

	err = run(ctx, model, filter, username)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(ctx context.Context, modelName string, filter vcsretriever.FilterReponse, username string) error {
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return fmt.Errorf("when ollama.New: %w", err)
	}

	repos, err := vcsretriever.GetRepos(username)
	if err != nil {
		return fmt.Errorf("erreur getRepos: %w", err)
	}

	// Pour chaque projet, récupérer et afficher le README
	for _, repo := range repos {
		log.Printf("trying to describe %+v\n", repo)
		description, err := vcsretriever.DescribeGitHubRepo(ctx, llm, repo, filter)
		if err != nil {
			fmt.Printf("when describeGitHubRepo: %v\n", err)
			continue
		}
		fmt.Printf("Description generated: %s\n", description)
		fmt.Println("----------------------------------------")
	}
	return nil
}
