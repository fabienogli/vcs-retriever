package main

import (
	"context"
	"fmt"
	"log"
	"os"

	vcsretriever "github.com/fabienogli/vcs-retriever"
	"github.com/fabienogli/vcs-retriever/github"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	ghFlag = "gh-username"
)

type arg struct {
	githubUsername string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vcs-retriever-cli",
	Short: "VCS Retriever will retrieve github projects",
	// 	Long: `PDL (Parallel Downloader) is a CLI library that will download an URL.

	// It will chunk the file in several files
	// `,
	// Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		githubUsername, err := cmd.Flags().GetString(ghFlag)
		if err != nil {
			return fmt.Errorf("err getting timeout %w", err)
		}
		if githubUsername == "" {
			return fmt.Errorf("⚠️ githubUsername not defined")
		}
		err = run(cmd.Context(), arg{
			githubUsername: githubUsername,
		})

		return err
	},
}

func init() {
	rootCmd.PersistentFlags().String(ghFlag, "", "github username")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, args arg) error {
	err := runGithub(ctx, args.githubUsername)
	if err != nil {
		return err
	}
	return nil

	err = runWithAI(ctx, args.githubUsername)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return nil
}

func runWithAI(ctx context.Context, username string) error {
	//specific to deepseek
	filter, err := vcsretriever.DeepseekFilter()
	if err != nil {
		return fmt.Errorf("when deepseekFilter: %w", err)
	}
	modelName := "deepseek-r1:1.5b"
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return fmt.Errorf("when ollama.New: %w", err)
	}

	return vcsretriever.SummarizeRepos(ctx, username, llm, filter)
}

func runGithub(ctx context.Context, username string) error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("when godotenv.Load: %w", err)
	}

	token := os.Getenv("GITHUB_TOKEN")

	if token == "" {
		return fmt.Errorf("⚠️ le token GitHub n'est pas défini !")
	}

	pinnedResponse, err := github.GetPinnedRepositories(token, username)

	if err != nil {
		return fmt.Errorf("when vcsretriever.GetPinnedRepositories: %w", err)
	}

	repos := vcsretriever.FromPinned(pinnedResponse, username)

	// Print response
	fmt.Println("Pinned Repositories:")
	for _, node := range repos {
		fmt.Println("node: ", node)
		readme, err := vcsretriever.ReadmeToByte(node)
		if err != nil {
			continue
		}
		fmt.Println(string(readme))
		return nil
	}
	return nil
}
