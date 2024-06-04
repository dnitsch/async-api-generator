const libReport = require('istanbul-lib-report')
const { createCoverageMap, createFileCoverage } = require('istanbul-lib-coverage')
const reports = require('istanbul-reports')
const mergedRaw = require('./.combined-raw.json')

let map = createCoverageMap()

for (const fk of Object.keys(mergedRaw)) {
  const reportKey = fk // .includes("/libs/") ? fk.split("/libs/")[1] : fk.split("/tasks/")[1]
  const jsonCoverageMap =  createCoverageMap({[reportKey]: createFileCoverage(mergedRaw[fk])});
  map.merge(jsonCoverageMap);
}

// create a context for report generation
const context = libReport.createContext({
  dir: './.coverage',
  defaultSummarizer: "nested",
  coverageMap: map, //.files().forEach((f) => )
// this is the map which we generated in above snippet
})

// create an instance of the relevant report class, passing the
// report name e.g. json/html/html-spa/text
const reportHtml = reports.create('html')

const reportJunit = reports.create('cobertura')
// call execute to synchronously create and write the report to disk
reportHtml.execute(context)

reportJunit.execute(context)
