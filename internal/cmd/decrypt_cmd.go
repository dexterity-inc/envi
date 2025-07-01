package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/tui"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Decrypt command flags
var (
	decryptInputFile     string
	decryptOutputFile    string
	decryptSelfContained bool
	decryptPassword      string
)

// decryptCmd is the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt encrypted environment variables",
	Long:  `Decrypt encrypted .env files or self-contained encrypted shares.`,
	Run:   runDecryptCommand,
}

// InitDecryptCommand sets up the decrypt command and its subcommands
func InitDecryptCommand() {
	// Initialize the command flags
	decryptCmd.Flags().StringVarP(&decryptInputFile, "input", "i", "", "Input file to decrypt")
	decryptCmd.Flags().StringVarP(&decryptOutputFile, "output", "o", ".env", "Output file path")
	decryptCmd.Flags().BoolVar(&decryptSelfContained, "self-contained", false, "Decrypt a self-contained encrypted share")
	decryptCmd.Flags().StringVarP(&decryptPassword, "password", "p", "", "Password for decryption")

	// Add encryption flags for standard decryption
	decryptCmd.Flags().BoolVar(&encryption.UseKeyFile, "use-key-file", false, "Use key file instead of password")
	decryptCmd.Flags().StringVarP(&encryption.EncryptionKeyFile, "key-file", "k", ".envi.key", "Path to encryption key file")
	decryptCmd.Flags().BoolVarP(&encryption.UseMaskedEncryption, "mask", "m", false, "Decrypt masked values (keep keys visible)")

	// Add the decrypt command to the root command
	rootCmd.AddCommand(decryptCmd)
}

// runDecryptCommand handles the decrypt command execution
func runDecryptCommand(cmd *cobra.Command, args []string) {
	// Handle self-contained decryption
	if decryptSelfContained {
		handleSelfContainedDecrypt(cmd)
		return
	}

	// Handle standard decryption
	handleStandardDecrypt(cmd)
}

// handleSelfContainedDecrypt handles self-contained encrypted share decryption
func handleSelfContainedDecrypt(cmd *cobra.Command) {
	// Get the share password
	var sharePassword string
	if decryptPassword != "" {
		sharePassword = decryptPassword
	} else {
		// Prompt for password
		if encryption.UseTUI {
			var err error
			sharePassword, err = tui.GetPassword("Enter password for self-contained share", false)
			if err != nil {
				utils.Error("Error getting password: %s", err)
				utils.Fatal("Failed to get password")
			}
		} else {
			utils.Info("Enter password for self-contained share: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				utils.Error("Error reading password: %s", err)
				utils.Fatal("Failed to read password")
			}
			utils.Info("")
			sharePassword = string(passwordBytes)
		}
	}

	if sharePassword == "" {
		utils.Error("Password cannot be empty")
		utils.Fatal("Invalid password")
	}

	// Read the input file
	if decryptInputFile == "" {
		utils.Error("Input file is required for self-contained decryption")
		utils.Info("Use --input to specify the file containing the encrypted content")
		utils.Fatal("Missing input file")
	}

	encryptedContent, err := os.ReadFile(decryptInputFile)
	if err != nil {
		utils.Error("Error reading input file: %s", err)
		utils.Fatal("Failed to read input file")
	}

	// Check if it's a self-contained share
	if !encryption.IsSelfContainedShare(encryptedContent) {
		utils.Error("The provided content is not a self-contained encrypted share")
		utils.Fatal("Invalid content format")
	}

	// Decrypt the content
	utils.Info("Decrypting self-contained share...")
	decryptedContent, err := encryption.DecryptSelfContainedShare(encryptedContent, sharePassword)
	if err != nil {
		utils.Error("Error decrypting content: %s", err)
		utils.Info("Please check that:")
		utils.Info("1. The password is correct")
		utils.Info("2. The encrypted content was copied completely")
		utils.Info("3. The content hasn't been modified")
		utils.Fatal("Decryption failed")
	}

	// Write the decrypted content
	err = os.WriteFile(decryptOutputFile, decryptedContent, 0600)
	if err != nil {
		utils.Error("Error writing output file: %s", err)
		utils.Fatal("Failed to write output file")
	}

	utils.Success("Successfully decrypted content to %s", decryptOutputFile)
}

// handleStandardDecrypt handles standard encrypted content decryption
func handleStandardDecrypt(cmd *cobra.Command) {
	// Read the input file
	if decryptInputFile == "" {
		utils.Error("Input file is required")
		utils.Info("Use --input to specify the file to decrypt")
		utils.Fatal("Missing input file")
	}

	encryptedContent, err := os.ReadFile(decryptInputFile)
	if err != nil {
		utils.Error("Error reading input file: %s", err)
		utils.Fatal("Failed to read input file")
	}

	// Check what type of encryption we're dealing with
	isEncrypted := encryption.IsEncrypted(encryptedContent)
	isMasked := encryption.IsMasked(encryptedContent)

	if !isEncrypted && !isMasked {
		utils.Error("The input file does not appear to be encrypted")
		utils.Info("Make sure you're using the correct decryption method")
		utils.Fatal("Invalid file format")
	}

	// Decrypt the content
	utils.Info("Decrypting content...")
	var decryptedContent []byte

	if isEncrypted {
		decryptedContent, err = encryption.DecryptContent(encryptedContent)
	} else if isMasked {
		decryptedContent, err = encryption.UnmaskEnvContent(encryptedContent)
	}

	if err != nil {
		utils.Error("Error decrypting content: %s", err)
		utils.Info("Please check your encryption key or password")
		utils.Fatal("Decryption failed")
	}

	// Write the decrypted content
	err = os.WriteFile(decryptOutputFile, decryptedContent, 0600)
	if err != nil {
		utils.Error("Error writing output file: %s", err)
		utils.Fatal("Failed to write output file")
	}

	utils.Success("Successfully decrypted content to %s", decryptOutputFile)
}
