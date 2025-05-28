package cmd

import (
	"fmt"

	"github.com/google/go-github/v37/github"
	"github.com/dexterity-inc/envi/internal/encryption"
)

// This file contains utility functions for the cmd package 

// createSharingReadmeContent generates README content for shared Gists
func createSharingReadmeContent(user *github.User, recipientUsername string, keyFilePath string) string {
	readmeContent := fmt.Sprintf("# Shared Environment Variables\n\n")
	readmeContent += fmt.Sprintf("This Gist contains environment variables shared by @%s with @%s.\n\n", *user.Login, recipientUsername)
	readmeContent += "## Instructions\n\n"
	readmeContent += "1. Click on the `.env` file above to view the shared environment variables\n"
	readmeContent += "2. Copy the contents to your local `.env` file\n"
	
	if encryption.UseEncryption || encryption.UseMaskedEncryption {
		readmeContent += "\n## Decryption Instructions\n\n"
		readmeContent += "These environment variables are encrypted. To use them, you will need:\n\n"
		
		if keyFilePath != "" {
			readmeContent += "1. The key file that was shared with you separately\n"
			readmeContent += "2. The envi tool installed on your machine\n\n"
			readmeContent += "To decrypt and use this file, run the following command:\n\n"
			readmeContent += "```shell\n"
			readmeContent += "# Replace <gist-id> with this Gist's ID and <keyfile> with the path to the key file\n"
			readmeContent += "envi pull -i <gist-id> --unmask --use-key-file --key-file <keyfile>\n"
			readmeContent += "```\n"
		} else {
			readmeContent += "1. The encryption password (request this from the sender)\n"
			readmeContent += "2. The envi tool installed on your machine\n\n"
			readmeContent += "To decrypt and use this file, run the following command:\n\n"
			readmeContent += "```shell\n"
			readmeContent += "# Replace <gist-id> with this Gist's ID\n"
			readmeContent += "envi pull -i <gist-id> --unmask\n"
			readmeContent += "```\n"
			readmeContent += "You will be prompted to enter the encryption password.\n"
		}
	}
	
	readmeContent += "\n\n## Getting started with envi\n\n"
	readmeContent += "If you don't have envi installed yet:\n\n"
	readmeContent += "```shell\n"
	readmeContent += "# macOS/Linux\n"
	readmeContent += "brew tap dexterity-inc/tap\n"
	readmeContent += "brew install envi\n\n"
	readmeContent += "# Windows\n"
	readmeContent += "scoop bucket add dexterity-inc https://github.com/dexterity-inc/scoop-bucket\n"
	readmeContent += "scoop install envi\n"
	readmeContent += "```\n"
	
	readmeContent += "\n\n---\n"
	readmeContent += "Shared using [envi](https://github.com/dexterity-inc/envi), an open-source environment variable manager"
	
	return readmeContent
} 