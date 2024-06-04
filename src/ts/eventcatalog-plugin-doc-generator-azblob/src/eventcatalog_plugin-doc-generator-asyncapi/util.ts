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
  // ###BEGIN_EVENTCATALOG_EXAMPLES###
  // #->eyJuYW1lIjoiUGFja2luZ0FyZWFFdmVudC0wLWV4YW1wbGUiLCJzdW1tYXJ5Ijoie1wiZmlsZVwiOlwic29tZWV2ZW50cG9jby5jc1sxMy0zMV1cIixcInBhdGhcIjpcIi9Vc2Vycy9kdXNhbm5pdHNjaG5laWRlci9naXQvbmV4dC93YXJlaG91c2UtaW50ZWdyYXRpb24vV2hkcy5Ub29scy5Bc3luY0FQSUdlbmVyYXRvci9zcmMvZ28vYXN5bmMtYXBpLWdlbi1kb2MvdGVzdC93aGRzLnNhbXBsZS9zcmMvc29tZWV2ZW50cG9jby5jc1wifSIsInBheWxvYWQiOiJuYW1lc3BhY2UgV2hkcy5QYWNraW5nLlNlcnZpY2VzLlBhY2tBcmVhLkNvbnRyYWN0cy5FdmVudHM7XG5cbnB1YmxpYyBjbGFzcyBQYWNraW5nQXJlYUV2ZW50IDogV2hkc01lc3NhZ2VcdTAwM2NQYWNraW5nQXJlYUV2ZW50UGF5bG9hZFx1MDAzZVxue1xuICAgIHB1YmxpYyBQYWNraW5nQXJlYUV2ZW50KFBhY2tpbmdBcmVhRXZlbnRQYXlsb2FkIHBheWxvYWQpXG4gICAge1xuICAgICAgICBNZXNzYWdlVHlwZU5hbWUgPSBuYW1lb2YoUGFja2luZ0FyZWFFdmVudCk7XG4gICAgICAgIFNvdXJjZVN5c3RlbSA9IFBhY2tBcmVhU2VydmljZUNvbnN0YW50cy5OYW1lO1xuICAgICAgICBHdWlkID0gR3VpZC5OZXdHdWlkKCk7XG4gICAgICAgIENyZWF0aW9uRGF0ZSA9IERhdGVUaW1lLlV0Y05vdztcbiAgICAgICAgTnVtYmVyID0gMTtcbiAgICAgICAgTnVtYmVyT2YgPSAxO1xuICAgICAgICBPd25lciA9IHN0cmluZy5FbXB0eTtcbiAgICAgICAgU3RyZWFtID0gU3RyaW5nLkVtcHR5O1xuICAgICAgICBQYXlsb2FkID0gcGF5bG9hZDtcbiAgICB9XG59In0=
  // ###END_EVENTCATALOG_EXAMPLES###
  // The output will look like the below
  // let output = {"name":"N6_BulkOrderConfiguration_v1-0-example","summary":"{\"file\":\"BulkOrderUpdatedEvent.cs[9-16]\",\"path\":\"/some/path/domain.Dft.Sorter6.Mapper/src/domain.Dft.Sorter6.Mapper.Models/WarehouseEvents/Version1/OrderUpdatedEvents/BulkOrderUpdatedEvent.cs\"}","payload":"namespace domain.Dft.Sorter6.Mapper.Models.WarehouseEvents.Version1.OrderUpdatedEvents\n{\n    public class BulkOrderUpdatedEvent : OrderUpdatedEvent\n    {\n    }\n}"}

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
