import fs from 'node:fs'
import path from 'node:path'

export function createReport({ scenario, logger }) {
  return {
    scenario,
    started_at: new Date().toISOString(),
    ended_at: null,
    duration_ms: null,
    start: { dimension: 'unknown', pos: null },
    end: null,
    checks: [],
    logs: logger.logs,
    result: 'fail',
    error: null,
  }
}

export function addCheck(report, name, pass, detail = {}) {
  report.checks.push({ name, pass, ...detail })
}

export function finalizeReport(report, { end, result, error }) {
  const ended = Date.now()
  report.ended_at = new Date(ended).toISOString()
  report.duration_ms = ended - Date.parse(report.started_at)
  report.end = end
  report.result = result
  report.error = error || null
}

export function writeReport(reportDir, report) {
  fs.mkdirSync(reportDir, { recursive: true })
  const ts = report.started_at.replace(/[:.]/g, '-')
  const file = `${ts}-${report.scenario}.json`
  const target = path.join(reportDir, file)
  fs.writeFileSync(target, `${JSON.stringify(report, null, 2)}\n`, 'utf8')
  return target
}
