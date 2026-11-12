import { useEffect, useState } from 'react'
import { api } from '../../shared/api.js'

// Loads what the current user did with issues and lets a page switch one
// interaction on or off.
export function useInteractions() {
  const [recorded, setRecorded] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    api('/users/me/interactions')
      .then(history => setRecorded(history.map(({ issueId, type }) => ({ issueId, type }))))
      .catch(failure => setError(failure.message))
  }, [])

  function has(issueId, type) {
    return recorded.some(item => item.issueId === issueId && item.type === type)
  }

  async function toggle(issueId, type) {
    try {
      if (has(issueId, type)) {
        await api(`/interactions/${issueId}/${type}`, { method: 'DELETE' })
        setRecorded(recorded.filter(item => !(item.issueId === issueId && item.type === type)))
      } else {
        await api('/interactions', { method: 'POST', body: { issueId, type } })
        setRecorded([...recorded, { issueId, type }])
      }
      setError('')
    } catch (failure) {
      setError(failure.message)
    }
  }

  return { has, toggle, error }
}
