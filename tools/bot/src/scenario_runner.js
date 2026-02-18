import { snapshotPosition } from './bot.js'
import { loadPortals } from './portals.js'
import { addCheck, createReport, finalizeReport, writeReport } from './report.js'
import { scenarios } from './scenarios/index.js'

export async function runScenarioOnce({ name, bot, logger, reportDir, portalsPath }) {
  const scenario = scenarios[name]
  if (!scenario) {
    throw new Error(`unknown scenario: ${name}`)
  }

  const portals = loadPortals(portalsPath)
  const report = createReport({ scenario: name, logger })
  report.start = { dimension: 'overworld', pos: snapshotPosition(bot) }

  try {
    const result = await scenario({
      bot,
      portals,
      log: logger.log,
      report,
      addCheck,
    })
    const passed = report.checks.every((c) => c.pass)
    finalizeReport(report, {
      end: { dimension: 'overworld', pos: result?.end || snapshotPosition(bot) },
      result: passed ? 'pass' : 'fail',
      error: null,
    })
    const file = writeReport(reportDir, report)
    return { report, file, passed }
  } catch (err) {
    finalizeReport(report, {
      end: { dimension: 'overworld', pos: snapshotPosition(bot) },
      result: 'fail',
      error: err.message,
    })
    const file = writeReport(reportDir, report)
    return { report, file, passed: false, error: err }
  }
}
