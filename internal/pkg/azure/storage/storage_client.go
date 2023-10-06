package storage

import (
	"bytes"
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type FileClient interface {
	Upload(file File) error
	GetList(group string) ([]File, error)
	GetFile(group string, name string) (*File, error)
	DeleteGroup(group string) error
}

type azureStorageClient struct {
	client *azblob.Client
}

func NewAzureStorageClient() FileClient {
	// https://learn.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-go?tabs=roles-azure-portal
	// TODO: replace <storage-account-name> with your actual storage account name
	url := "https://sobrian.blob.core.windows.net/"
	// ctx := context.Background()

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Println(err.Error())

		return nil
	}

	client, err := azblob.NewClient(url, credential, nil)
	if err != nil {
		log.Println(err.Error())

		return nil
	}

	return &azureStorageClient{client: client}
}

func (c *azureStorageClient) Upload(file File) error {
	c.client.CreateContainer(context.Background(), file.Group, &azblob.CreateContainerOptions{})

	res, err := c.client.UploadStream(context.Background(), file.Group, file.Name, file.Content, &azblob.UploadStreamOptions{})
	if err != nil {
		log.Println(err.Error())

		return err
	}

	log.Println(res.ContentMD5)
	return nil
}

func (c *azureStorageClient) GetList(group string) ([]File, error) {
	pager := c.client.NewListBlobsFlatPager(group, nil)
	var files []File
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}

		for _, blob := range page.Segment.BlobItems {
			log.Println(*blob.Name)
			files = append(files, File{
				Group:      group,
				Name:       *blob.Name,
				UploadDate: *blob.Properties.LastModified,
			})
		}
	}

	return files, nil
}

func (c *azureStorageClient) GetFile(group string, name string) (*File, error) {
	// Download the blob
	ctx := context.Background()
	res, err := c.client.DownloadStream(ctx, group, name, nil)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	// Read the blob
	downloadedData := bytes.Buffer{}
	retryReader := res.NewRetryReader(ctx, &azblob.RetryReaderOptions{})
	_, err = downloadedData.ReadFrom(retryReader)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	// Close the blob
	err = retryReader.Close()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &File{
		Group:      group,
		Name:       name,
		Content:    &downloadedData,
		UploadDate: *res.LastModified,
	}, nil
}

func (c *azureStorageClient) DeleteGroup(group string) error {
	_, err := c.client.DeleteContainer(context.Background(), group, nil)
	if err != nil {
		return err
	}

	return nil
}
