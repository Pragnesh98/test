package speechlogging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type azureBlobStorage struct {
	StorageAccount string `json:"storage_account"`
	ContainerName  string `json:"container_name"`
}

type audioStoreMetadata struct {
	AzureBlobStorage azureBlobStorage `json:"azure_blob_storage"`
}

// ContainerURLFmt specifies the URL format azure blob store
const ContainerURLFmt = "https://%s.blob.core.windows.net/%s"

type azureBlobStore struct {
	accountName   string
	containerName string
	credential    azblob.Credential
	pathPrefix    string
}

// NewAzureUploader returns an instance of azure upload implementation.
func NewAzureUploader(accountName string, accessKey string, containerName string, prefix string) (*azureBlobStore, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accessKey)

	if err != nil {
		return nil, err
	}
	return &azureBlobStore{
		accountName:   accountName,
		containerName: containerName,
		credential:    credential,
		pathPrefix:    prefix,
	}, nil
}

func (s *azureBlobStore) GetInfo() string {
	info := audioStoreMetadata{
		AzureBlobStorage: azureBlobStorage{
			ContainerName:  s.containerName,
			StorageAccount: s.accountName,
		},
	}

	result, err := json.Marshal(info)
	if err != nil {
		log.Printf("Failed to marshal azure store info %s\n", err)
		result = []byte{}
	}
	return string(result)
}

func (s *azureBlobStore) getFullName(name string) string {
	return path.Join(s.pathPrefix, name)
}

func (s *azureBlobStore) Download(ctx context.Context, name string) (body io.ReadCloser, err error) {
	containerURL, err := url.Parse(fmt.Sprintf(ContainerURLFmt, s.accountName, s.containerName))
	p := azblob.NewPipeline(s.credential, azblob.PipelineOptions{})
	azContainer := azblob.NewContainerURL(*containerURL, p)
	blobURL := azContainer.NewBlobURL(s.getFullName(name))
	downloadResp, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return
	}
	body = downloadResp.Body(azblob.RetryReaderOptions{})
	return
}

func (s *azureBlobStore) Upload(ctx context.Context, name string, data []byte) (URL string, err error) {
	containerURL, err := url.Parse(
		fmt.Sprintf(ContainerURLFmt, s.accountName, s.containerName))

	p := azblob.NewPipeline(s.credential, azblob.PipelineOptions{})

	azContainer := azblob.NewContainerURL(*containerURL, p)

	log.Println(s.getFullName(name))
	blockBlob := azContainer.NewBlockBlobURL(s.getFullName(name))
	log.Println(blockBlob.BlobURL)
	_, err = azblob.UploadBufferToBlockBlob(ctx, data, blockBlob, azblob.UploadToBlockBlobOptions{})

	log.Println(err)
	if err != nil {
		log.Println(err)
	} else {
		URL = blockBlob.BlobURL.String()
	}
	return
}

// ListBlobs reads a list of blobs from a container and writes it to the writer.
func (s *azureBlobStore) ListBlobs(ctx context.Context, prefix string, w io.WriteCloser) error {

	defer w.Close()

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", s.accountName, s.containerName))

	p := azblob.NewPipeline(s.credential, azblob.PipelineOptions{})

	containerURL := azblob.NewContainerURL(*URL, p)
	options := azblob.ListBlobsSegmentOptions{MaxResults: 1024, Prefix: s.getFullName(prefix)}
	marker := azblob.Marker{}
	for marker.NotDone() {
		listBlobsResponse, err := containerURL.ListBlobsFlatSegment(ctx, marker, options)
		if err != nil {
			fmt.Printf("Error while listing blobs %s", err)
			return err
		}
		marker = listBlobsResponse.NextMarker
		for _, blobItem := range listBlobsResponse.Segment.BlobItems {
			_, err := fmt.Fprintln(w, blobItem.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *azureBlobStore) DeleteBlob(ctx context.Context, name string) error {
	containerURL, err := url.Parse(
		fmt.Sprintf(ContainerURLFmt, s.accountName, s.containerName))

	p := azblob.NewPipeline(s.credential, azblob.PipelineOptions{})

	azContainer := azblob.NewContainerURL(*containerURL, p)

	log.Println(s.getFullName(name))
	blockBlob := azContainer.NewBlockBlobURL(s.getFullName(name))
	log.Println(blockBlob.BlobURL)

	_, err = blockBlob.Delete(ctx, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	if err != nil {
		log.Println(err)
	}

	return err
}
