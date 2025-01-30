package vcsretriever

import (
	"context"
	"fmt"
	"regexp"

	"github.com/tmc/langchaingo/llms"
)

func DeepseekFilter() (FilterReponse, error) {
	// Regex pour extraire tout ce qui suit la balise </think>
	re, err := regexp.Compile(`(</think>)(.*)`)
	if err != nil {
		return nil, fmt.Errorf("when regexp.Compile: %w", err)
	}
	return func(s string) string {
		// Trouver le match
		match := re.FindStringSubmatch(s)

		// Si un match est trouvé, afficher le texte extrait
		if len(match) > 1 {
			return match[1]
		}
		return s
	}, nil
}

func generatePrompt(repo Repo) string {
	// Créer un prompt pour décrire le repository en utilisant le modèle LLM
	return fmt.Sprintf(`
You are an assistant that help to describe a Github repository, which are files written in software developpement language.
Your goal is to describe best the project. The most import thing of this project is the Readme, which is the introduction to the project.
Here is the Readme:
%s
----------------------------
The title %q and the description %q could help but are less important than the Readme above.
Your description should be limited to 100 characters.`, repo.Readme, repo.FullName, repo.Description)
}

// Fonction pour interroger un modèle LLM via Ollama
func queryOllamaLLM(ctx context.Context, model llms.Model, prompt string) (string, error) {
	resp, err := model.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: prompt,
				},
			},
		},
	},
	// llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
	// 	fmt.Print(string(chunk))
	// 	return nil
	// }),
	)
	if err != nil {
		return "", fmt.Errorf("when model.GenerateContent")
	}
	if len(resp.Choices) < 1 {
		return "", fmt.Errorf("resp choices below 0")
	}

	// Retourner la réponse générée par l'LLM
	return resp.Choices[0].Content, nil
}
