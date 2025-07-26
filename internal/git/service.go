package git

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/context/git_component"
	"github.com/cjlapao/locally-cli/internal/executer"
	"github.com/cjlapao/locally-cli/internal/icons"

	"github.com/cjlapao/common-go/helper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var globalGitService *GitService

type GitService struct{}

func New() *GitService {
	config := configuration.Get()
	svc := GitService{}

	sources := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.SOURCES_PATH)

	if !helper.DirectoryExists(sources) {
		if helper.CreateDirectory(sources, fs.ModePerm) {
			notify.InfoWithIcon(icons.IconCheckMark, "Sources folder created")
		} else {
			notify.Critical("There was an error creating the sources folder")
		}
	}

	return &svc
}

func Get() *GitService {
	if globalGitService != nil {
		return globalGitService
	}

	return New()
}

func (svc *GitService) Clone(source, destination string, cleanBeforeClone bool) error {
	return svc.CloneWithCredentials(source, destination, nil, cleanBeforeClone)
}

func (svc *GitService) CloneWithCredentials(source, destination string, gitCredentials *git_component.GitCredentials, cleanBeforeClone bool) error {
	var publicKey *ssh.PublicKeys
	config := configuration.Get()
	sourceUrl, err := url.Parse(source)
	if err != nil {
		return err
	}

	sources := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.SOURCES_PATH)

	if !helper.DirectoryExists(sources) {
		if helper.CreateDirectory(sources, fs.ModePerm) {
			notify.InfoWithIcon(icons.IconCheckMark, "Folder %s created", sources)
		} else {
			notify.Critical("There was an error creating the %s folder", sources)
		}
	}

	if cleanBeforeClone {
		if err := svc.clean(destination); err != nil {
			return err
		}
	}

	source, publicKey, err = InsertCredentials(sourceUrl, gitCredentials)
	if err != nil {
		return err
	}

	sourceFileCount := 0
	if helper.DirectoryExists(destination) {
		files, _ := ioutil.ReadDir(destination)
		sourceFileCount = len(files)
		if sourceFileCount == 0 {
			if err := helper.DeleteFile(destination); err != nil {
				return err
			}
		}
	}

	if helper.DirectoryExists(destination) && sourceFileCount > 0 {
		notify.Info("Destination folder %s already exists, changing to master and getting latest", destination)
		currentFolder, changeDirErr := os.Getwd()
		if changeDirErr != nil {
			return changeDirErr
		}
		changeDirErr = os.Chdir(destination)
		if changeDirErr != nil {
			return changeDirErr
		}

		runArgs := make([]string, 0)
		runArgs = append(runArgs, "pull")
		if common.IsDebug() {
			notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
		}

		output, err := executer.ExecuteWithNoOutput("git", runArgs...)

		changeDirErr = os.Chdir(currentFolder)
		if changeDirErr != nil {
			return changeDirErr
		}

		if err != nil {
			notify.FromError(err, "Something wrong running git pull on master")
			if output.GetAllOutput() != "" {
				notify.Error(output.GetAllOutput())
			}
			return err
		}
		if common.IsDebug() {
			notify.Debug("Output: %s", output.GetAllOutput())
		}
	} else {
		notify.Info("Starting to clone %s to %s", sourceUrl, destination)
		_, err = git.PlainClone(destination, false, &git.CloneOptions{
			URL:      source,
			Progress: os.Stdout,
			Auth:     publicKey,
		})
	}

	return err
}

func getPrivateKey(source string) (*ssh.PublicKeys, error) {
	if !helper.FileExists(source) {
		return nil, fmt.Errorf("file %s was not found", source)
	}

	fileInfo, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		notify.Debug("User added a folder path instead of the file, attempting the default key name")
		source = helper.JoinPath(source, "id_ed25519")
		if !helper.FileExists(source) {
			return nil, fmt.Errorf("file %s was not found", source)
		}
	}

	b, err := helper.ReadFromFile(source)
	if err != nil {
		return nil, err
	}

	var publicKey *ssh.PublicKeys
	publicKey, keyError := ssh.NewPublicKeys("git", b, "")
	if keyError != nil {
		return nil, keyError
	}

	return publicKey, nil
}

func InsertCredentials(sourceUrl *url.URL, gitCredentials *git_component.GitCredentials) (string, *ssh.PublicKeys, error) {
	if gitCredentials == nil {
		return sourceUrl.String(), nil, nil
	}
	notify.Debug("Inserting Credentials into url %s", sourceUrl.String())
	var publicKey *ssh.PublicKeys
	var err error

	source := sourceUrl.String()
	accessToken := gitCredentials.AccessToken
	username := gitCredentials.Username
	password := gitCredentials.Password
	privateKeyPath := gitCredentials.PrivateKeyPath

	if privateKeyPath != "" {
		source = strings.ReplaceAll(strings.ReplaceAll(sourceUrl.String(), "https://github.com/", "git@github.com:"), "http://github.com/", "git@github.com:")
		publicKey, err = getPrivateKey(privateKeyPath)
		if err != nil {
			return source, nil, err
		}
		notify.Debug("Found ssh key, changing the url to use the access token, %s", source)
	}

	if accessToken != "" {
		sourceUrl.User = url.UserPassword("oauth2", accessToken)
		source = sourceUrl.String()
		notify.Debug("Found access token, changing the url to use the access token, %s", source)
	}

	if username != "" && password != "" {
		sourceUrl.User = url.UserPassword(username, password)
		source = sourceUrl.String()
		notify.Debug("Found username/password, changing the url to use the access token, %s", source)
	}

	return source, publicKey, nil
}

func (svc *GitService) clean(source string) error {
	if helper.DirectoryExists(source) {
		notify.Debug("Cleaning the existing folder %s", source)
		if err := helper.DeleteAllFiles(source); err != nil {
			return err
		}

		if err := helper.DeleteFile(source); err != nil {
			return err
		}
	}

	return nil
}
