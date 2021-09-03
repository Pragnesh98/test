// +build integration

package speechlogging

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/google/uuid"
)

func IntegrationTestAzureUpload(t *testing.T) {
	az := azureBlobStore{
		accountName:   "logblobstore",
		containerName: "test-stt-log-documents",
	}

	key := os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	credential, err := azblob.NewSharedKeyCredential(az.accountName, key)

	if err != nil {
		t.Fatalf("Failed to create credentials %s", err)
	}

	az.credential = credential

	randomName := uuid.New()

	t.Log("Uploading blob")
	u, err := az.Upload(context.Background(), randomName.String(), []byte("test"))

	if err != nil {
		t.Fatalf("Failed to uplaod blob")
	}

	defer func() {
		t.Log("Deleting blob")
		az.DeleteBlob(context.Background(), randomName.String())
		t.Log("Deleted blob")
	}()

	expectedURL := fmt.Sprintf(ContainerURLFmt, az.accountName, az.containerName+"/"+randomName.String())
	if u != expectedURL {
		t.Fatalf("Expected url %q, got url %q", expectedURL, u)
	}

	t.Log("Downloading blob")
	body, err := az.Download(context.Background(), randomName.String())

	if err != nil {
		t.Fatalf("Failed to download blob %s, %s", randomName.String(), err)
	}

	contents, err := ioutil.ReadAll(body)
	body.Close()

	if string(contents) != "test" {
		t.Fatalf("Expected %q, got %q", "test", contents)
	}
}
