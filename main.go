// TTW Software Team
// Mathis Van Eetvelde
// 2021-present

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

const (
	scope            = "https://www.googleapis.com/auth/drive.file"
	filenameInput    = "filename"
	nameInput        = "name"
	folderIdInput    = "folderId"
	credentialsInput = "credentials"
	overwriteInput   = "overwrite"
)

func main() {

	// get filename argument from action input
	filename := githubactions.GetInput(filenameInput)
	if filename == "" {
		missingInput(filenameInput)
	}

	// get name argument from action input
	name := githubactions.GetInput(nameInput)

	// get folderId argument from action input
	folderId := githubactions.GetInput(folderIdInput)
	if folderId == "" {
		missingInput(folderIdInput)
	}

	// get base64 encoded credentials argument from action input
	credentials := githubactions.GetInput(credentialsInput)
	if credentials == "" {
		missingInput(credentialsInput)
	}
	// add base64 encoded credentials argument to mask
	githubactions.AddMask(credentials)

	// decode credentials to []byte
	decodedCredentials, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		githubactions.Fatalf(fmt.Sprintf("base64 decoding of 'credentials' failed with error: %v", err))
	}

	creds := strings.TrimSuffix(string(decodedCredentials), "\n")

	// add decoded credentials argument to mask
	githubactions.AddMask(creds)

	// fetching a JWT config with credentials and the right scope
	conf, err := google.JWTConfigFromJSON([]byte(creds), scope)
	if err != nil {
		githubactions.Fatalf(fmt.Sprintf("fetching JWT credentials failed with error: %v", err))
	}

	// instantiating a new drive service
	ctx := context.Background()
	svc, err := drive.New(conf.Client(ctx))
	if err != nil {
		log.Println(err)
	}

	file, err := os.Open(filename)
	if err != nil {
		githubactions.Fatalf(fmt.Sprintf("opening file with filename: %v failed with error: %v", filename, err))
	}

	// decide name of file in GDrive
	if name == "" {
		name = file.Name()
	}

	f := &drive.File{
		Name:    name,
		Parents: []string{folderId},
	}

	if overwrite := githubactions.GetInput(overwriteInput); overwrite != "" {
		overwriteFlag, err := strconv.ParseBool(overwrite)
		if err != nil {
			errorInput(overwriteInput, err)
		}
		if overwriteFlag {
			r, err := svc.Files.List().Do()
			if err != nil {
				githubactions.Fatalf("list files failed with error: %v", err)
			}
			for _, i := range r.Files {
				if filename == i.Name {
					if err := svc.Files.Delete(i.Id).Do(); err != nil {
						githubactions.Fatalf("delete file: %+v failed with error: %v", i.Id, err)
					}
				}
			}
		}
	}

	dst, err := svc.Files.Create(f).Media(file).Do()
	if err != nil {
		githubactions.Fatalf(fmt.Sprintf("creating file: %+v failed with error: %v", f, err))
	}

	githubactions.SetOutput("downloadURL",
		fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", dst.Id))
}

func missingInput(inputName string) {
	githubactions.Fatalf(fmt.Sprintf("missing input '%v'", inputName))
}

func errorInput(inputName string, err error) {
	githubactions.Fatalf(fmt.Sprintf("error input '%v': %s", inputName, err))
}
