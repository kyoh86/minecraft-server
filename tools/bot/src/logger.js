export function createLogger(options = {}) {
  const logs = []
  const useStderr = options.stderr === true

  function log(message, extra = {}) {
    const entry = {
      at: new Date().toISOString(),
      message,
      ...extra,
    }
    logs.push(entry)
    const suffix = Object.keys(extra).length ? ` ${JSON.stringify(extra)}` : ''
    const line = `[bot] ${entry.at} ${message}${suffix}\n`
    if (useStderr) {
      process.stderr.write(line)
      return
    }
    process.stdout.write(line)
  }

  return { logs, log }
}
