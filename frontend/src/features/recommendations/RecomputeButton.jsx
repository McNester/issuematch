import { useState } from 'react'
import { api } from '../../shared/api.js'

// Lets an admin run the similarity job without waiting for its schedule.
export default function RecomputeButton() {
  const [status, setStatus] = useState('')
  const [running, setRunning] = useState(false)

  async function recompute() {
    setRunning(true)
    setStatus('')
    try {
      const { issues, durationMs } = await api('/admin/recompute', { method: 'POST' })
      setStatus(`Updated ${issues.toLocaleString('en-US')} issues in ${durationMs.toLocaleString('en-US')} ms`)
    } catch (failure) {
      setStatus(failure.message)
    }
    setRunning(false)
  }

  return (
    <>
      {status && <span className="muted">{status}</span>}
      <button type="button" className="button-secondary" disabled={running} onClick={recompute}>
        {running ? 'Recomputing…' : 'Recompute similarities'}
      </button>
    </>
  )
}
