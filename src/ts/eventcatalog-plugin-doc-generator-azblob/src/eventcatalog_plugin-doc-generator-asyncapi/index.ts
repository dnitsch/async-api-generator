// Currently borrowed from
// https://github.com/boyney123/eventcatalog/tree/master/packages/eventcatalog-plugin-generator-asyncapi
// a few changes have been made
// - updated to latest `@asyncapi/parser`
// - added an examples
// TODO: PR a change upstream

import {
  AsyncAPIDocumentInterface as AsyncAPIDocument,
  Input,
  Parser,
} from "@asyncapi/parser"
import type {
  Domain,
  LoadContext,
  Service,
} from "@eventcatalog/types"
import utils from "@eventcatalog/utils"
import fs from "fs-extra"
import path, { join } from "path"

import { EventCatalogMsgExamples, ParsedEvent, enrichServiceIndexMarkdown, eventMsgBuilder, extractEventCatalogExamples } from "./util"

export type AsyncAPIPluginOptions = {
  pathToSpec: string | string[];
  versionEvents?: boolean;
  externalAsyncAPIUrl?: string;
  renderMermaidDiagram?: boolean;
  renderNodeGraph?: boolean;
  domainName?: string;
  domainSummary?: string;
};

export interface AsyncAPIConvertRemoteOptions extends AsyncAPIPluginOptions {
  outputDir: string;
  writer: NodeJS.WriteStream;
}

const getServiceFromAsyncDoc = (doc: AsyncAPIDocument): Service => {
  let tags = doc.info().tags();
  const url = tags.find((t) => t.name() === "repoUrl")?.description();
  const language = tags.find((t) => t.name() === "repoLang")?.description();
  return {
    name: doc.info().title(),
    summary: doc.info().description() || "",
    repository: { url, language },
  };
};

const getAllEventsFromAsyncDoc = (
  doc: AsyncAPIDocument,
  parsedExamples: EventCatalogMsgExamples,
  service: Service,
  options: AsyncAPIConvertRemoteOptions
): ParsedEvent[] => {
  const { externalAsyncAPIUrl } = options;

  const allMessages = handleChannelMessages(
    doc,
    parsedExamples,
    service,
    externalAsyncAPIUrl
  );
  // the same service can be the producer and consumer of events, check and merge any matchs.
  const uniqueMessages = allMessages.reduce((acc: any, message: any) => {
    const messageAlreadyDefined = acc.findIndex(
      (m: any) => m.name === message.name
    );

    if (messageAlreadyDefined > -1) {
      acc[messageAlreadyDefined] = { ...acc[messageAlreadyDefined], message };
    } else {
      acc.push(message);
    }
    return acc;
  }, []);

  return uniqueMessages;
};



const handleChannelMessages = (
  doc: AsyncAPIDocument,
  parsedExamples: EventCatalogMsgExamples,
  service: Service,
  externalAsyncAPIUrl?: string
): ParsedEvent[] => {
  // const channels = doc.allChannels()
  const allMessages = doc
    .allChannels()
    .all()
    .reduce((data: any, channel) => {
      let eventsFromMessages: ParsedEvent[] = [] as ParsedEvent[];
      // channels should map one to one to operations
      for (const operation of channel?.operations()?.all()) {
        for (const message of operation.messages()) {
          let messageName = message.name(); // ||

          // If no name can be found from the message, and AsyncAPI defaults to "anonymous" value, try get the name from the payload itself
          if (message.name()?.includes("anonymous-")) {
            messageName = message.payload()?.$id() || messageName;
          }

          const schema = message.payload()?.json();
          const externalLink = {
            label: `View event in AsyncAPI`,
            url: `${externalAsyncAPIUrl}#message-${messageName}`,
          };

          let asyncAPIExamples = Object.keys(parsedExamples)
            .map((name: string) => ({ ...parsedExamples[name] }))
            .filter((f) => f.name === message.id());
          
          eventsFromMessages = [
            ...eventsFromMessages,
            eventMsgBuilder({
              name: messageName as string,
              message,
              operation,
              version: doc.info().version(),
              externalLink,
              examples: asyncAPIExamples.flatMap((_, idx) => ({
                  fileName: asyncAPIExamples[idx].summary.file,
                  fileContent: asyncAPIExamples[idx].payload,
                })),
              schema,
              service,
              channelNotes: channel.description() || "",
            }),
          ];
        }
      }
      return data.concat(eventsFromMessages);
    }, []);

  return allMessages;
};

const parseAsyncAPIFile = async (
  pathToFile: string,
  options: AsyncAPIConvertRemoteOptions,
  copyFrontMatter: boolean
) => {
  const {
    versionEvents = true,
    renderMermaidDiagram = false,
    renderNodeGraph = true,
    domainName = "",
    domainSummary = "",
    outputDir,
  } = options;

  let asyncAPIFile: Input = {} as Input;

  asyncAPIFile = await fs
    .readFile(pathToFile, "utf-8")
    // .then((d) => {
    //   asyncAPIFile = d;
    // })
    .catch((error: any) => {
      console.error(error);
      throw new Error(`Failed to read file with provided path`);
    });

  // {validateOptions: {allowedSeverity: {warning: true, error: false}}}
  // use default validation options
  const doc = await new Parser().parse(asyncAPIFile, { 
    parseSchemas: false, 
    validateOptions: { 
      allowedSeverity: {
        warning: true, error: false, hint: true, info: true
      }
    }}).catch((ex) => {
    throw ex;
  });

  if (!doc.document) {
    console.error(`File (%s) failed to load`, pathToFile)
    throw new Error("unable to parse AsyncAPI document");
  }

  const parsedExamples = extractEventCatalogExamples(asyncAPIFile as string);

  const service = getServiceFromAsyncDoc(doc?.document as AsyncAPIDocument);
  const events = getAllEventsFromAsyncDoc(
    doc?.document as AsyncAPIDocument,
    parsedExamples,
    service,
    options
  );

  if (!outputDir) {
    throw new Error("Please provide catalog url (env variable PROJECT_DIR)");
  }

  if (domainName) {
    const { writeDomainToCatalog } = utils({ catalogDirectory: outputDir });

    const domain: Domain = {
      name: domainName,
      summary: domainSummary,
    };

    await writeDomainToCatalog(domain, {
      useMarkdownContentFromExistingDomain: true,
      renderMermaidDiagram,
      renderNodeGraph,
    });
  }

  const { writeServiceToCatalog, writeEventToCatalog } = utils({
    catalogDirectory: domainName
      ? path.join(outputDir, "domains", domainName)
      : outputDir,
  });

  // Note: this is a temporary workaround 
  // until an upstream PR is made to allow for a more dynamic content building of MD pages in EventCatalog
  const svcDesc = service.summary;
  service.summary = "";
  const servicePath = await writeServiceToCatalog(service, {
    useMarkdownContentFromExistingService: false,
    renderMermaidDiagram,
    renderNodeGraph,
  });

  // now we pass the previously capture description 
  // into the extended markdown builder to preserve it 
  // in the complete output
  await enrichServiceIndexMarkdown(join(servicePath.path, "index.md"), svcDesc);

  const eventFiles = events.map(async (event: ParsedEvent) => {
    // As this is adapted from an existing lib which hasn't made proper use of types
    // It seems that the producers/consumers on Event are defined as Service
    // however they have to be a string for the writeEventToCatalog to work
    const { schema, ...eventData } = event as any;

    await writeEventToCatalog(eventData, {
      useMarkdownContentFromExistingEvent: false,
      versionExistingEvent: versionEvents,
      renderMermaidDiagram,
      renderNodeGraph,
      frontMatterToCopyToNewVersions: {
        // only do consumers and producers if its not the first file.
        consumers: false, // copyFrontMatter,
        producers: false, // copyFrontMatter,
      },
      markdownContent: `
${eventData.description}

<NodeGraph />

### Code Example(s)

<EventExamples />

<Schema />
`,
      codeExamples: eventData?.examples,
      schema: {
        extension: "json",
        fileContent: schema,
      },
    });
  });

  // write all events to folders
  Promise.all(eventFiles);

  return {
    generatedEvents: events,
  };
};

/**
 * main
 */
export default async (
  _: LoadContext,
  options: AsyncAPIConvertRemoteOptions
) => {
  const { pathToSpec } = options;

  const listOfAsyncAPIFilesToParse = Array.isArray(pathToSpec)
    ? pathToSpec
    : [pathToSpec];

  if (listOfAsyncAPIFilesToParse.length === 0 || !pathToSpec) {
    throw new Error("No file provided in plugin.");
  }

  // on first parse of files don't copy any frontmatter over.
  const parsers = listOfAsyncAPIFilesToParse.map((specFile, index) =>
    parseAsyncAPIFile(specFile, options, index !== 0)
  );

  const data = await Promise.all(parsers);
  const totalEvents = data.reduce(
    (sum, { generatedEvents }) => sum + generatedEvents.length,
    0
  );

  options.writer.write(`\x1b[32m 
    Successfully parsed ${listOfAsyncAPIFilesToParse.length} AsyncAPI file/s. 
    Generated ${totalEvents} events \x1b[0m`);
};
