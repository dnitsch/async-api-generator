import { PageSettings, PagedAsyncIterableIterator } from "@azure/core-paging/types/latest/core-paging"
import { BlobDownloadResponseParsed, BlobItem, ContainerListBlobFlatSegmentResponse, ContainerListBlobsOptions } from "@azure/storage-blob/types/latest/storage-blob"
import { MakeDirectoryOptions, PathLike } from "fs"
import { readdirSync, rm } from "fs-extra"
import { tmpdir } from "os"
import path from "path"
import { cwd } from "process"
import { PathMutationOptions, PluginOptions } from "."
import { BlobIface, ContainerIface } from "./client"
import * as mockAsyncApiPlugin from "./eventcatalog_plugin-doc-generator-asyncapi"
import { FSIface, configByDomain, convertToEventCatalog, downloadDomains, parsePath } from "./worker"

jest.setTimeout(60000);

describe("worker", () => {

  let mockDownload: BlobDownloadResponseParsed;

  let mockBlobClient: BlobIface;

  let mockContainerClient: ContainerIface;

  let fsClientIMock: FSIface;

  let pluginOpts: PluginOptions;
  let outputPath = path.join(tmpdir(), "mock-out-test")
  
    describe("configuration by domain", () => {
    
      test.each([
        // case 1
        [["urn:biz_domain:bounded_ctx_domain1:service1","urn:biz_domain:bounded_ctx_domain1:service2","urn:biz_domain:bounded_ctx_domain1:service3","urn:biz_domain:bounded_ctx_domain1:service4","urn:biz_domain:bounded_ctx_domain1:service5","urn:biz_domain:bounded_ctx_domain2:service1","urn:biz_domain:bounded_ctx_domain2:service2","urn:biz_domain:bounded_ctx_domain2:service3","urn:biz_domain:bounded_ctx_domain2:service4","urn:biz_domain:bounded_ctx_domain3:service5"],
          3, [0, 5]],
        [["urn:biz_domain:bounded_ctx_domain1:service1","urn:biz_domain:bounded_ctx_domain1:service2","urn:biz_domain:bounded_ctx_domain1:service3","urn:biz_domain:bounded_ctx_domain1:service4","urn:biz_domain:bounded_ctx_domain1:service5","urn:biz_domain:bounded_ctx_domain2:service1","urn:biz_domain:bounded_ctx_domain2:service2","urn:biz_domain:bounded_ctx_domain2:service3","urn:biz_domain:bounded_ctx_domain2:service4","urn:biz_domain:bounded_ctx_domain3:service5"],
          3, [1, 4]],
        [["urn:biz_domain:bounded_ctx_domain1:service1","urn:biz_domain:bounded_ctx_domain1:service2","urn:biz_domain:bounded_ctx_domain1:service3","urn:biz_domain:bounded_ctx_domain1:service4","urn:biz_domain:bounded_ctx_domain1:service5","urn:biz_domain:bounded_ctx_domain2:service1","urn:biz_domain:bounded_ctx_domain2:service2","urn:biz_domain:bounded_ctx_domain2:service3","urn:biz_domain:bounded_ctx_domain2:service4","urn:biz_domain:bounded_ctx_domain3:service5"],
          3, [2, 1]],
        [["urn:biz_domain:bounded_ctx_domain1:service1","urn:biz_domain:bounded_ctx_domain1:service2",
        "urn:biz_domain:bounded_ctx_domain1:service3","urn:biz_domain:bounded_ctx_domain2:service1"],
          2, [0, 3]],
        [["urn:biz_domain:bounded_ctx_domain1:service1","urn:biz_domain:bounded_ctx_domain1:service2",
          "urn:biz_domain:bounded_ctx_domain1:service3","urn:biz_domain:bounded_ctx_domain2:service1"],
            2, [1, 1]],
      ])("asyncAPI serviceId (%s), expects bounded_ctx_domains: %d at these indexs: %s", async (paths: string[], bounded_ctx_domains: number, pathSpec: number[]) => {
        // Arrange
        const [ pathIndex, pathLength ] = pathSpec
        
        // Act
        const got = configByDomain({ generatedPaths: paths, outputDir: "./foo", writer:  process.stdout})
        
        // Assert
        expect(got.length).toBe(bounded_ctx_domains)
        expect(got[pathIndex].pathToSpec.length).toBe(pathLength)
      })
    })
    describe("generate AsyncAPI docs by domain", () => {
      test("succeeds with sample inputs", async () => {
        // Arrange
        let err = null 
        const generatedPaths = ["urn:domain:s2s:service1.yml","urn:domain:s2s:service2.yml"].map((p) => path.join(cwd(), "src", "__mocks__","sample", p))
        const baseOutDir = path.join(cwd(),".ignore")
        const writer = process.stdout, outputDir = baseOutDir

        // Act
        await convertToEventCatalog({ generatedPaths, writer, outputDir })
          .catch((ex) => {
            err = ex
            console.log('ex :>> ', ex.message, ex.stack);
          })

        // Assert
        expect(err).toBe(null)
        // Assert contents correct
        let domainEventsOut = path.join(baseOutDir,"domains","s2s","events")
        let eventsDir = readdirSync(domainEventsOut)
        
        expect(eventsDir.length).toBe(2)
        let testEventDir = readdirSync(path.join(domainEventsOut,"BizContextAreaEvent"),{})
        expect(testEventDir.length).toBe(2)
        testEventDir.map((t) => expect(["index.md","schema.json"].find((f) => f == t )).not.toBe(null))
        expect(testEventDir)
        rm(baseOutDir, {recursive: true, force: true})
      })

      test("errors on AsyncAPI plugin", async () => {
        // Arrange
        let err = null 
        const generatedPaths = ["urn:domain:s2s:service1.yml","urn:domain:s2s:service2.yml"].map((p) => path.join(cwd(), "src", "__mocks__","sample", p))
        const baseOutDir = path.join(cwd(),".ignore")
        const writer = process.stdout, outputDir = baseOutDir
        jest.spyOn(mockAsyncApiPlugin, "default").mockRejectedValueOnce(new Error("async api conversion error"))
        // Act
        await convertToEventCatalog({ generatedPaths, writer, outputDir })
          .catch((ex) => {
            err = ex
          })

        // Assert
        expect(err).not.toBe(null)
      })
    })
    describe("parsePath", () => {
      test.each([
        ["without additional slash", "/home/foo/domains/foo/bar.cs","delete/this/foo/bar.cs"],
        ["with additional slash", "/home/foo/domains/foo/bar.cs", "delete/this//foo/bar.cs"]
      ])
      ("should match all %s", (desc, ttexpect, ttblob ) => {
        const opts = {
          baseDir: "/home/foo",
          find: "delete/this",
          replace: "domains",
        } as PathMutationOptions;
  
        expect(parsePath(opts, ttblob)).toBe(ttexpect);
      });
    });

    describe("downloadDomains", () => {
      beforeEach(() => {
        // global set env
        process.env.PROJECT_DIR = "foo";
        // fsclient 
        fsClientIMock = {
          existsSync: () => true,
          mkdir: async (path, opts) => "foo",
        } as FSIface
    
        // plugin opts
        pluginOpts = {
          blobAccount: "__test___",
          blobContainer: "eventcatalog",
          keyToDestinationMutation: {
            baseDir: "bar",
            replace: "tada",
            find: "bux"
          }
        } as PluginOptions;
        mockBlobClient = {
          downloadToFile(
            filePath,
            offset,
            count,
            options
          ): Promise<BlobDownloadResponseParsed> {
            return Promise.resolve(mockDownload);
          },
        };
        mockContainerClient = {
          listBlobsFlat(
            options?: ContainerListBlobsOptions
          ): PagedAsyncIterableIterator<
            BlobItem,
            ContainerListBlobFlatSegmentResponse
          > {
            let itemOne = {} as BlobItem;
            itemOne.name = "foo.txt"
            let byPageMock = {} as ContainerListBlobFlatSegmentResponse;
            let count = 0;
            async function* listSegments(
              options: any = {}
            ): AsyncIterableIterator<ContainerListBlobFlatSegmentResponse> {
              do {
                for (const page of [[byPageMock]]) {
                  count++;
                  yield* page;
                }
              } while (count < 1);
            }
    
            async function* listItems(
              options: any = {}
            ): AsyncIterableIterator<BlobItem> {
              count++
              if (count < 2) {
                  yield* [itemOne]
              }
              return Promise.resolve({done: true})
            }
    
            return {
              next() {
                return listItems().next();
              },
              [Symbol.asyncIterator]() {
                return this;
              },
              byPage: (settings: PageSettings = {}) => {
                return listSegments();
              },
            };
          },
          getBlockBlobClient(blobName: string): BlobIface {
            return mockBlobClient;
          },
        };
      });
      afterEach(async () => {
        await rm(outputPath, { recursive: true, force: true}).catch((ex) => {
          console.log('ignore empty dir delete :>> ', ex.message);
        })
      });
    
      test("succeeds with correct input", async () => {
        // Arrange
        let err = null;
  
        // Act
        const generatedPaths = await downloadDomains(
          mockContainerClient, fsClientIMock, pluginOpts).catch((e) => {
          err = e;
        });
  
        // Assert
        expect(err).toBe(null);
        expect(generatedPaths).toBeInstanceOf(Array)
        expect((generatedPaths as string[])?.length).toBe(1)
      });
  
      test("errors in a filesystem operation", async () => {
        // Arrange
        let err = null;
        fsClientIMock.existsSync = (path: PathLike) => false
        fsClientIMock.mkdir = async (path: PathLike, options: MakeDirectoryOptions) => {
           return Promise.reject(new Error("unable to make dir"))
        }
  
        // Act
        await downloadDomains(
          mockContainerClient, fsClientIMock, pluginOpts).catch((e) => {
          err = e;
        });
        // Assert
        expect(err).not.toBe(null);
        expect(err).toBeInstanceOf(Error);
      })
  
      test("errors in remote Client operation downloading a file", async () => {
        // Arrange
        let err = null
        
        jest.spyOn(mockBlobClient, "downloadToFile").mockRejectedValueOnce(new Error("unable to download")) 
        
        // Act
        await downloadDomains(
          mockContainerClient, fsClientIMock, pluginOpts).catch((e) => {
          err = e;
        });
        
        // Assert
        expect(err).not.toBe(null);
        expect(err).toBeInstanceOf(Error);
      })
    });
})
