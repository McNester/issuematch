import { useEffect, useState } from 'react'
import { api } from '../../shared/api.js'
import IssueRow from '../issues/IssueRow.jsx'

export default function RecommendationsPage() {
  const [items, setItems] = useState([])
  const [bookmarked, setBookmarked] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    api('/recommendations')
      .then(response => setItems(response.items))
      .catch(failure => setError(failure.message))
  }, [])

  async function bookmark(issueId) {
    try {
      await api('/interactions', { method: 'POST', body: { issueId, type: 'bookmark' } })
      setBookmarked([...bookmarked, issueId])
    } catch (failure) {
      setError(failure.message)
    }
  }

  return (
    <main className="page">
      <h1 className="page-title">For you</h1>
      {error && <p className="error">{error}</p>}
      <div className="list">
        {items.map(({ issue, basedOn }) => (
          <IssueRow key={issue.id} issue={issue} bookmarked={bookmarked.includes(issue.id)} onBookmark={bookmark}
            note={basedOn && 'Similar to: ' + basedOn} />
        ))}
      </div>
    </main>
  )
}
