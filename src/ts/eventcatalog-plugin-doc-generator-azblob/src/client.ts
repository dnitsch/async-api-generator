import { PagedAsyncIterableIterator } from "@azure/core-paging"
import { DefaultAzureCredential } from "@azure/identity"
import {
  BlobDownloadOptions,
  BlobDownloadResponseParsed,
  BlobItem,
  BlobServiceClient,
  ContainerClient,
  ContainerListBlobFlatSegmentResponse,
  ContainerListBlobsOptions,
} from "@azure/storage-blob"

export interface ClientIface {
  credential: () => () => DefaultAzureCredential;
  blobClient: (cred: DefaultAzureCredential, account: string) => ContainerIface;
}

export interface ContainerIface {
  listBlobsFlat: (
    options?: ContainerListBlobsOptions
  ) => PagedAsyncIterableIterator<
    BlobItem,
    ContainerListBlobFlatSegmentResponse
  >;
  getBlockBlobClient: (blobName: string) => BlobIface;
}

export interface BlobIface {
  downloadToFile: (
    filePath: string,
    offset?: number,
    count?: number,
    options?: BlobDownloadOptions
  ) => Promise<BlobDownloadResponseParsed>;
}

/**
 * 
 * @param credential 
 * @param blobAccount 
 * @param blobContainer 
 */
export function blobClient(credential: DefaultAzureCredential, blobAccount: string): BlobServiceClient {
    return new BlobServiceClient(
        `https://${blobAccount}.blob.core.windows.net`,
        credential
      );
}

/**
 * 
 * @param client 
 * @param blobContainer 
 * @returns 
 */
export function containerClient(client: BlobServiceClient, blobContainer: string): ContainerClient {
    return client.getContainerClient(blobContainer);
}
