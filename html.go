package vcsretriever

import (
	"fmt"
	"os"
)

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
