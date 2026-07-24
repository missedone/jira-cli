package download

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Download saves an attachment from a Jira issue.`
	examples = `$ jira issue attachment download ISSUE-1 10001

# Save the attachment to a specific path
$ jira issue attachment download ISSUE-1 10001 --output ./downloads/report.pdf`
	filePerm = 0o644
)

// NewCmdDownload is an attachment download command.
func NewCmdDownload() *cobra.Command {
	cmd := cobra.Command{
		Use:     "download ISSUE-KEY ATTACHMENT-ID",
		Short:   "Download an issue attachment",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"dl"},
		Annotations: map[string]string{
			"help:args": `ISSUE-KEY	Issue key, eg: ISSUE-1
ATTACHMENT-ID	Numeric Jira attachment ID, eg: 10001`,
		},
		Args: cobra.ExactArgs(2),
		RunE: run,
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (defaults to the attachment filename)")

	return &cmd
}

func run(cmd *cobra.Command, args []string) error {
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	attachmentID := args[1]
	client := api.DefaultClient(debug)

	metadata, err := func() (*jira.Attachment, error) {
		s := cmdutil.Info("Fetching attachment details...")
		defer s.Stop()

		return api.ProxyGetIssueAttachment(client, issueKey, attachmentID)
	}()
	if err != nil {
		return err
	}

	destination, err := outputPath(output, metadata.Filename)
	if err != nil {
		return err
	}

	err = writeNewFile(destination, func(dst io.Writer) error {
		s := cmdutil.Info(fmt.Sprintf("Downloading %q...", metadata.Filename))
		defer s.Stop()

		return api.ProxyDownloadAttachment(client, attachmentID, dst)
	})
	if err != nil {
		return err
	}

	absDestination, err := filepath.Abs(destination)
	if err != nil {
		absDestination = destination
	}
	cmdutil.Success("Attachment downloaded to %s", absDestination)

	return nil
}

func outputPath(output, filename string) (string, error) {
	filename = filepath.Base(filename)
	if filename == "." || filename == string(filepath.Separator) || filename == "" {
		return "", fmt.Errorf("jira: attachment has an invalid filename")
	}
	if output == "" {
		return filename, nil
	}

	info, err := os.Stat(output)
	if err == nil && info.IsDir() {
		return filepath.Join(output, filename), nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return output, nil
}

func writeNewFile(path string, download func(io.Writer) error) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if err != nil {
		return err
	}

	if err := download(file); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}
