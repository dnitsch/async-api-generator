import { BlobDownloadOptions, BlobItem } from "@azure/storage-blob"
import { LoadContext } from "@eventcatalog/types"
import { MakeDirectoryOptions, PathLike } from "fs"
import path from "path"
import { PathMutationOptions, PluginOptions } from "."
import { ContainerIface } from "./client"
import asyncApiPlugin, {
  AsyncAPIConvertRemoteOptions
} from "./eventcatalog_plugin-doc-generator-asyncapi"

/**
 *
 * @param containerClient
 */
export const downloadDomains = async (
  containerClient: ContainerIface,
  fsClient: FSIface,
  opts: PluginOptions
): Promise<string[]> => {
  let tasks = [] as Promise<string>[];
  for await (const blob of containerClient.listBlobsFlat()) {
    let newTask = async (
      blob: BlobItem,
      opts: PluginOptions
    ): Promise<string> => {
      console.debug(blob.name);
      const blobClient = containerClient.getBlockBlobClient(blob.name)

      let wrPath = parsePath(opts.keyToDestinationMutation, blob.name);
      await ensureDir(wrPath, fsClient);

      await blobClient
        .downloadToFile(wrPath, 0, undefined, {} as BlobDownloadOptions)
        .catch((ex) => {
          console.error("errored on: %s, ex: %s", blob.name, ex.message);
          throw ex;
        });
      console.debug("successfully updated: %s", blob.name);
      return wrPath;
    };
    tasks.push(newTask(blob, opts));
  }

  return await Promise.all(tasks).catch((ex) => {
    throw ex;
  });
};

/**
 * FS interface for easy mocking
 */
export interface FSIface {
  existsSync: (path: PathLike) => boolean;
  mkdir: (
    path: PathLike,
    options: MakeDirectoryOptions
  ) => Promise<string | undefined>;
}

/**
 * ensureDir either creates or returns if dir exists
 * accepts a file from blob pager
 * @param file
 * @param fs
 */
export async function ensureDir(file: string, fs: FSIface): Promise<void> {
  const dir = path.dirname(file);
  if (!fs.existsSync(dir)) {
    await fs.mkdir(dir, { recursive: true });
  }
}

export function parsePath(options: PathMutationOptions, blob: string): string {
  const { baseDir, find, replace } = options;
  console.debug(`baseDir: ${baseDir}, find: ${find}, replace: ${replace}`);
  return path.join(baseDir, blob.replace(find, replace));
}

export const convertToEventCatalog = async ({
  generatedPaths,
  outputDir,
  writer,
}: {
  generatedPaths: string[];
  outputDir: string;
  writer: NodeJS.WriteStream;
}): Promise<void> => {
  // as the plugin has all it needs in a single file
  // these can all be run in parallel
  // let tasks = [] as Promise<string>[];
  for (const domainOpt of configByDomain({ generatedPaths, outputDir, writer })) {
    await asyncApiPlugin(
      {} as LoadContext,
      domainOpt
    ).catch((ex: Error) => {
      // stop on first error for now...
      throw ex;
    });

  }
};

/**
 * 
 * @param generatedPaths
 * @param outputDir
 * @param writer
 * @returns
 */
export const configByDomain = (
{ generatedPaths, outputDir, writer }: { generatedPaths: string[]; outputDir: string; writer: NodeJS.WriteStream; }): AsyncAPIConvertRemoteOptions[] => {
  // split by domain
  const domainPaths = generatedPaths.map((p, _idx, _arr) => {
    //  since all the AsyncAPI docs are generated and stored under the following name pattern
    //  urn:BIZ_DOMAIN:BOUNDED_CONTEXT_DOMAIN:SERVICE_NAME
    //  we want to extract the `BOUNDED_CONTEXT_DOMAIN` and assign it to the opts object as the name of the domain
    //  this will ensure that events are organized by domain
    return { domain: p.split(":")[2], file_path: p };
  });

  let optBase = {
    versionEvents: true,
    outputDir,
    writer,
  } as AsyncAPIConvertRemoteOptions;

  const optsByDomain = domainPaths.reduce(
    (acc: AsyncAPIConvertRemoteOptions[], d) => {
      const foundIdx = acc.findIndex((a) => a.domainName === d.domain);
      if (foundIdx > -1) {
        // acc[foundIdx] =
        acc[foundIdx].pathToSpec = [...acc[foundIdx].pathToSpec, d.file_path];
        return acc;
      }
      return [
        ...acc,
        { ...optBase, domainName: d.domain, pathToSpec: [d.file_path] } as AsyncAPIConvertRemoteOptions,
      ];
    },
    []
  );

  return optsByDomain; 
};
