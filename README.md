# VCS-RETRIEVER 

## Project Overview  
This repository is designed to enable users to explore github profile and to summarize their github projects.

##  Key Features  

###  Retrievieng all the repositories from a user  
###  LLM Capabilities - using LLM to summarize the project

## Usage
Run
```bash
go run cmd/vcs-retriever-cli/cli.go --help
```
## Roadmap 
- Summarize the project only using the readme (without AI)
- - you should add blacklist/whitelist to avoid/select which repositories to exclude/include
- Summarize the project by providing the readme to the LLM model
- Summarize the project by using 