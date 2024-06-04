# AsyncAPI GenDoc Plugin

AsyncAPI Gendoc plugin for EventCatalog, should be used inside the `eventcatalog.config.js` along with the build process.

`npm run generate` will execute the plugin to autopopulate the domains directory.

```javascript
module.exports = {
  // existing eventcatalog.config.js
  // ...
  // 
  generators: [
    [
      '@dnitsch/plugin-doc-generator-asyncapi-remote-source',
      {
        blobAccount: "ACCOUNT_NAME",  
        blobContainer:  "eventcatalog", 
        outputDir: ".", // will use the current directory in eventcatalog
        keyToDestinationMutation: {
          baseDir: process.env.PROJECT_DIR,
          find: "domains/domains",
          replace: "domains"
        }
      },
    ],
  ],
};
```

The plugin will first download the AsyncAPI documents from a specified remote location, currently this is from AZureBlobStorage account and container.

It will then write the files to a temp location and invoke a call to a modified asyncapi-doc-generator to create all the eventcatalog artifacts by domain.

> this is currently hardcoded
