package azure

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/model"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

func UploadRecording(
	ctx context.Context,
	callSID string,
	channelID string,
	botID string,
	filePath string,
) (string, error) {

	// Wait for some time before uploading the recording so that full file can be written
	time.Sleep(3 * time.Second)

	containerInfo, err := getRecordingBlobstoreDetail(callSID, channelID, botID)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to get recording storage container info. Error: [%#v]", err)
		return "", err
	}

	switch containerInfo.Mode {
	case "SAS":
		blobURL, err := toBlobWithSASToken(ctx, callSID, channelID, containerInfo, filePath)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to upload the file to blob storage. Error: [%#v]", err)
			return "", err
		}
		return blobURL, nil
	default:
		blobURL, err := toBlobWithACCKey(ctx, callSID, channelID, containerInfo, filePath)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to upload the file to blob storage. Error: [%#v]", err)
			return "", err
		}
		return blobURL, nil
	}
}

func getRecordingBlobstoreDetail(
	callSID string,
	channelID string,
	botID string,
) (*model.BotBlobInfo, error) {
	botMappingInfo := configmanager.ConfStore.AzureRecordingStorageCustom

	for _, b := range botMappingInfo.BotMapings {
		if b.BotId == botID {
			ymlogger.LogInfof(callSID, "Using custom azure recording storage container")
			return &b, nil
		}
	}
	ymlogger.LogInfo(callSID, "Using default recording storage")

	containerInfo := &model.BotBlobInfo{
		AccountName:   os.Getenv("AZURE_STORAGE_ACCOUNT"),
		AccountKey:    os.Getenv("AZURE_STORAGE_ACCESS_KEY"),
		ContainerName: configmanager.ConfStore.AzureRecordingContainerName,
		Mode:          "ACCKEY",
		BotId:         botID,
	}
	if len(containerInfo.AccountName) == 0 || len(containerInfo.AccountKey) == 0 {
		ymlogger.LogError(callSID, "Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
		return nil, errors.New("Environment variables are not set")
	}
	return containerInfo, nil
}

func toBlobWithSASToken(
	ctx context.Context,
	callSID string,
	channelID string,
	containerInfo *model.BotBlobInfo,
	filePath string,
) (string, error) {
	fileEle := strings.Split(filePath, "/")
	fileName := fileEle[len(fileEle)-1]
	if len(containerInfo.BotId) > 0 {
		fileName = containerInfo.BotId + "/" +
			time.Now().Format("2006-01-02") + "/" +
			call.GetDialedNumber(channelID).WithZeroNationalFormat + "_" +
			time.Now().Format("2006-01-02T15:04:05") + "_" +
			fileName
	}

	endpoint := fmt.Sprintf("https://%s.blob.core.windows.net/", containerInfo.AccountName)
	client, err := storage.NewAccountSASClientFromEndpointToken(endpoint, containerInfo.SASToken)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while reading the file. Error: [%#v]", err)
	}

	blobClient := client.GetBlobService()
	containers, _ := blobClient.ListContainers(storage.ListContainersParameters{})
	ymlogger.LogInfof(callSID, "[%#v]", blobClient)
	ymlogger.LogInfof(callSID, "Error while reading the file. Error: [%#v]", err)
	if containers == nil {
		ymlogger.LogErrorf(callSID, "Error while reading the file. Error: [%#v]", err)
	}

	containerRef := blobClient.GetContainerReference(containerInfo.ContainerName)
	ymlogger.LogInfof(callSID, "Container Reference found:[%s]", containerRef.Name)

	_, err = containerRef.CreateIfNotExists(nil)
	if err != nil {
		ymlogger.LogError(callSID, "Couldn't create Block blob from file data")
		return "", err
	}

	blobRef := containerRef.GetBlobReference(fileName)
	f, _ := os.Open(filePath)
	err = blobRef.CreateBlockBlobFromReader(f, nil)
	if err != nil {
		ymlogger.LogError(callSID, "Couldn't create Block blob from file data")
		return "", err
	}
	ymlogger.LogInfof(callSID, "Uploaded file to blob: [%s]", blobRef.GetURL())
	return blobRef.GetURL(), nil
}

func toBlobWithACCKey(
	ctx context.Context,
	callSID string,
	channelID string,
	containerInfo *model.BotBlobInfo,
	filePath string,
) (string, error) {
	fileEle := strings.Split(filePath, "/")
	fileName := fileEle[len(fileEle)-1]
	if len(containerInfo.BotId) > 0 {
		fileName = containerInfo.BotId + "/" + time.Now().Format("2006-01-02") + "/" + fileName
	}

	u, err := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", containerInfo.AccountName, containerInfo.ContainerName, fileName))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to parse the URL. Error: [%#v]", err)
		return "", err
	}
	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(containerInfo.AccountName, containerInfo.AccountKey)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Invalid credentials. Error: [%#v]", err.Error())
		return "", err
	}

	blockBlobURL := azblob.NewBlockBlobURL(*u, azblob.NewPipeline(credential, azblob.PipelineOptions{}))
	// Read file contents with retries
	var dat []byte
	for i := 0; i < 3; i++ {
		dat, err = ioutil.ReadFile(filePath)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while reading the file. Error: [%#v]", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return "", err
	}

	// Specify any needed options to UploadToBlockBlobOptions (https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#UploadToBlockBlobOptions)
	o := azblob.UploadToBlockBlobOptions{
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{
			ContentType: "audio/wav",
		},
	}
	_, err = azblob.UploadBufferToBlockBlob(ctx, dat, blockBlobURL, o)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to upload the file to blob storage. Error: [%#v]", err)
		return "", err
	}
	ymlogger.LogInfof(callSID, "Uploaded file to blob: [%s]", blockBlobURL.String())
	return blockBlobURL.String(), nil
}
