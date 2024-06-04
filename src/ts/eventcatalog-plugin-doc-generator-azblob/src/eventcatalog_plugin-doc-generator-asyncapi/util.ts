import { MessageInterface, OperationInterface } from "@asyncapi/parser"
import type { Event, Service, Tag } from "@eventcatalog/types"
import { CodeExample } from "@eventcatalog/utils/lib/types"
import fs from "fs-extra"

export interface ParsedEvent extends Omit<Event, "producers" | "consumers"> {
    // export interface ParsedEvent extends Event {
    description: string;
    examples?: CodeExample[];
    producers: string[];
    consumers: string[];
}

export const eventMsgBuilder = ({
  name,
  message,
  operation,
  version,
  examples,
  schema,
  service,
  externalLink,
  channelNotes,
}: {
  name: string;
  message: MessageInterface;
  operation: OperationInterface;
  version: string;
  examples: CodeExample[];
  schema: any;
  service: Service;
  channelNotes: string;
  externalLink?: Tag;
}): ParsedEvent => {

  return {
    name: name,
    summary: message.summary(),
    description: `${channelNotes}\n${operation.description()}\n${message.description()}`,
    version: version,
    producers: operation.isSend() ? [service.name] : [],
    consumers: operation.isReceive() ? [service.name] : [],
    externalLinks: externalLink ? [externalLink] : [],
    schema: JSON.stringify(schema, null, 4),
    examples: examples,
    badges: [],
  };
};

export const BASE64_SUMMARY_MARKER = "#->";

export interface EventCatalogMsgExampleSummary {
  file: string;
  path: string;
}

export interface EventCatalogMsgExample {
  name: string;
  summary: EventCatalogMsgExampleSummary;
  payload: string;
}

export interface EventCatalogMsgExamples {
  [key: string]: EventCatalogMsgExample;
}

/**
 * extractEventCatalogExamples is a custom extension on top of the AsyncAPI generation
 * @description
 * input `rawDoc` may include
 * @param rawDoc
 * @returns
 * @example
 *
 */
export const extractEventCatalogExamples = (
  rawDoc: string
): EventCatalogMsgExamples => {

  let resp = {} as EventCatalogMsgExamples;

  for (const b64Summary of rawDoc.matchAll(/#->(.*$)/gm)) {
    if (b64Summary.length > 1) {
      let summary = Buffer.from(b64Summary[1].trimEnd(), "base64").toString(
        "utf-8"
      );
      const parsedSummary = JSON.parse(summary) as EventCatalogMsgExample;

      parsedSummary.summary = JSON.parse(
        parsedSummary.summary as unknown as string
      ) as EventCatalogMsgExampleSummary;
      parsedSummary.summary.file = parsedSummary.summary.file.split("[")[0];
      resp[parsedSummary.summary.file] = parsedSummary;
    }
  }

  return resp;
};


export const enrichServiceIndexMarkdown = async (
    serviceFilePath: string,
    svcDesc: string
  ): Promise<void> => {
    // NOTE: this is less than ideal way of doing
    // need to extend the utils writeServiceToCatalog
    // to accept markdownContent same as event.
    // 
    // Generating and reading from the same file 
    // could potentially see race condition under certain Disk I/O conditions
    const original = await fs
      .readFile(serviceFilePath, "utf-8")
      .catch((error: any) => {
        console.error(error);
        throw new Error(`Failed to read service file\n${error.message}`);
      });
  
    let yamlContent = original.split("---")[1];
    const mdContent = `---
${yamlContent}
---

${svcDesc}

<NodeGraph />
`;
  
    await fs
      .writeFile(serviceFilePath, mdContent, { encoding: "utf-8" })
      .catch((error: any) => {
        console.error(error);
        throw new Error(`Failed to write file with new service description\n${error.message}`);
      });
  };
