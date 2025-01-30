package vcsretriever

// Structure pour représenter un dépôt GitHub
type Repository struct {
	Name            string
	FullName        string
	Description     string
	URL             string
	User            string
	PrimaryLanguage string
	Readme          []byte
}
