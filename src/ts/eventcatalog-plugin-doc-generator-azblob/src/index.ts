import { DefaultAzureCredential } from "@azure/identity"
import type { LoadContext } from "@eventcatalog/types"
import { existsSync } from "fs"
import { mkdir } from "fs/promises"
import { blobClient, containerClient, ContainerIface } from "./client"
import { convertToEventCatalog, downloadDomains, FSIface } from "./worker"

/**
 * Options for the plugin
 *
 * example:
 * ```
 *  generators: [
 *   [
 *       'plugin-doc-generator-blob-container',
 *        {
 *       },
 *    ],
 *   ],
 * ```
 */
export interface PluginOptions {
  blobAccount: string;
  blobContainer: string;
  downloadDir?: string
  outputDir?: string
  // 
  keyToDestinationMutation: PathMutationOptions
}

export interface PathMutationOptions {
    baseDir: string
    find: string
    replace: string
}

/* istanbul ignore next */
/**
 * default func that gets called by the plugin invoker in eventcatalog
 */
export default async (_: LoadContext, options: PluginOptions) => {
  const { blobAccount, blobContainer, keyToDestinationMutation } = options;

  if (keyToDestinationMutation.baseDir.length < 1) {
    throw new Error("Please provide catalog url (env variable PROJECT_DIR) or hardcode into options");
  }

  const fsOp = {
    mkdir: mkdir,
    existsSync: existsSync,
  } as FSIface;
  
  const containerOp = containerClient(
    blobClient(new DefaultAzureCredential(), blobAccount),
    blobContainer
  )
  
  await generateFromRemoteSource(
    containerOp,
    fsOp,
    options,
    process.stdout
  )
};

export const generateFromRemoteSource = async (containerClient: ContainerIface, fsClient: FSIface, opts: PluginOptions, writer: NodeJS.WriteStream) => {
  const { blobAccount, blobContainer, outputDir } = opts
  writer.write(`\x1b[36m Beginning plugin \x1b[0m`)

  const generatedPaths = await downloadDomains(
    containerClient,
    fsClient,
    opts
  ).catch((ex) => {
      writer.write(`\x1b[31m failed to download processed asyncAPI docs: ${ex.message}. Unable to proceed further. \x1b[0m`)
    throw ex;
  });

  await convertToEventCatalog({generatedPaths, outputDir: outputDir as string, writer}).catch((ex) => {
    writer.write(`\x1b[31m ${ex.message} \x1b[0m`)
    throw ex;
  })
  writer.write(
    `\x1b[32m Succesfully downloaded generated AsyncAPI documents from "${blobAccount}/${blobContainer}". 
    Generated domains directory for EventCatalog. \x1b[0m`
  );
}
