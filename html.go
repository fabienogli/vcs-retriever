package vcsretriever

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fabienogli/vcs-retriever/github"
	"github.com/yuin/goldmark"
)

func ReadmeToByte(repo Repository) ([]byte, error) {
	readmeBody, err := github.RetrievingReadme(repo.User, repo.Name)
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

// Fonction pour écrire plusieurs contenus HTML dans un fichier
func writeHTMLToFile(filePath string, htmlContents [][]byte) error {
	// Ouvrir ou créer le fichier en mode écriture
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("erreur d'ouverture du fichier : %v", err)
	}
	defer file.Close()

	// Parcourir les contenus HTML et les écrire dans le fichier
	for _, content := range htmlContents {
		// Écrire chaque contenu HTML dans le fichier
		_, err := file.Write(content)
		if err != nil {
			return fmt.Errorf("erreur lors de l'écriture dans le fichier : %v", err)
		}
		// Ajouter une nouvelle ligne ou un séparateur entre les contenus si nécessaire
		_, err = file.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("erreur lors de l'ajout de saut de ligne dans le fichier : %v", err)
		}
	}

	return nil
}
