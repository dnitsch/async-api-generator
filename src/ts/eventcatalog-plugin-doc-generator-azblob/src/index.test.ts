import { LoadContext } from "@eventcatalog/types"
import { tmpdir } from "os"
import path from "path"
import { cwd } from "process"
import plugin, { PluginOptions, generateFromRemoteSource } from "."
import { ContainerIface } from "./client"
import * as mockWorker from "./worker"
import { FSIface } from "./worker"

jest.setTimeout(60000);

describe("index", () => {

  let mockContainerClient: ContainerIface;

  let fsClientIMock: FSIface;

  let pluginOpts: PluginOptions;
  let outputPath = path.join(tmpdir(), "mock-out-test")

  test.skip("integration test", async () => {
    let err = null
    await plugin(
      {} as LoadContext,
      {
        blobAccount: "stdevdneuntest",
        blobContainer: "processed",
        outputDir: ".ignore",
        keyToDestinationMutation: {
          baseDir: ".test/",
          find: "asyncapi",
          replace: "",
        },
      } as PluginOptions
    ).catch((ex) => {
      err = ex
    })
    expect(err).toBe(null) 
  });

  describe("generateFromSource", () => {
    beforeEach(() => {
      jest.spyOn(mockWorker, "downloadDomains").mockImplementation(async() => { return Promise.resolve(["path1", "path2"])})
      jest.spyOn(mockWorker, "convertToEventCatalog").mockImplementation(async() => { Promise.resolve({})})
      pluginOpts = {
        blobAccount: "__test___",
        blobContainer: "eventcatalog",
        keyToDestinationMutation: {
          baseDir: "bar",
          replace: "tada",
          find: "bux"
        }
      } as PluginOptions;
    })
    test("succeeds with correct inputs", async () => {
      // Arrange
      let err = null
      pluginOpts.keyToDestinationMutation = { 
        baseDir: cwd(), 
        find: "foo.txt", 
        replace: path.join("src", "__mocks__", "sample", "urn:domain:s2s:service1.yml")
      }
      pluginOpts.outputDir = outputPath
      // Act
      await generateFromRemoteSource(mockContainerClient, fsClientIMock, pluginOpts, process.stdout).catch((ex) => {
        err = ex
      })
      // Assert
      expect(err).toBe(null);
    })

    test("errors on download of generated AsyncAPI Docs", async () => {
      // Arrange
      let err = null

      jest.spyOn(mockWorker, "downloadDomains").mockRejectedValueOnce(new Error("failed to download from remote"))

      // Act
      await generateFromRemoteSource(mockContainerClient, fsClientIMock, pluginOpts, process.stdout).catch((ex) => {
        err = ex
      })

      // Assert
      expect(err).not.toBe(null)
    })

    test("errors on conversion of AsyncAPI to EventCatalog document layout", async () => {
      // Arrange
      let err = null

      pluginOpts.keyToDestinationMutation = { 
        baseDir: cwd(), 
        find: "foo.txt", 
        replace: path.join("src", "__mocks__","sample", "urn:domain:s2s:service1.yml")
      }
      pluginOpts.outputDir = outputPath
      
      // mock error as soon as possible
      jest.spyOn(mockWorker, "convertToEventCatalog").mockRejectedValueOnce(new Error("failed to convert"))

      // Act
      await generateFromRemoteSource(mockContainerClient, fsClientIMock, pluginOpts, process.stdout).catch((ex) => {
        err = ex
      })

      // Assert
      expect(err).not.toBe(null)
    })
  })

});
