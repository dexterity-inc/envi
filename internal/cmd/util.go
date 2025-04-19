package cmd

import (
	"fmt"

	"github.com/google/go-github/v37/github"
	"github.com/dexterity-inc/envi/internal/encryption"
)

// This file contains utility functions for the cmd package 

// createSharingReadmeContent generates README content for shared Gists
func createSharingReadmeContent(user *github.User, recipientUsername string) string {
	readmeContent := fmt.Sprintf("# Shared Environment Variables\n\n")
	readmeContent += fmt.Sprintf("This Gist contains environment variables shared by @%s with @%s.\n\n", *user.Login, recipientUsername)
	readmeContent += "## Instructions\n\n"
	readmeContent += "1. Click on the `.env` file above to view the shared environment variables\n"
	readmeContent += "2. Copy the contents to your local `.env` file\n"
	
	if encryption.UseTUI {
		readmeContent += "\nIf the content is encrypted, you'll need to request the decryption password from the sender."
	}
	
	readmeContent += "\n\n---\n"
	readmeContent += "Shared using [envi](https://github.com/dexterity-inc/envi), an open-source environment variable manager"
	
	return readmeContent
} 