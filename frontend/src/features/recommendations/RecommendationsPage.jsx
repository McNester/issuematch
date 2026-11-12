import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../shared/api.js'
import IssueRow from '../issues/IssueRow.jsx'
import { useInteractions } from '../issues/useInteractions.js'

export default function RecommendationsPage() {
  const [personal, setPersonal] = useState([])
  const [popular, setPopular] = useState([])
  const [languages, setLanguages] = useState([])
  const [error, setError] = useState('')
  const interactions = useInteractions()

  useEffect(() => {
    async function load() {
      try {
        const [recommended, user] = await Promise.all([api('/recommendations?limit=10'), api('/users/me')])
        setPersonal(recommended.items)
        setLanguages(user.languages)
        const trending = await api('/recommendations/trending?limit=20&languages=' + encodeURIComponent(user.languages.join(',')))
        const alreadyShown = recommended.items.map(item => item.issue.id)
        setPopular(trending.items.filter(item => !alreadyShown.includes(item.issue.id)).slice(0, 5))
      } catch (failure) {
        setError(failure.message)
      }
    }
    load()
  }, [])

  function rows(items) {
    return (
      <div className="list">
        {items.map(({ issue, basedOn }) => (
          <IssueRow key={issue.id} issue={issue} note={basedOn && 'Similar to: ' + basedOn}
            bookmarked={interactions.has(issue.id, 'bookmark')}
            onBookmark={issueId => interactions.toggle(issueId, 'bookmark')} />
        ))}
      </div>
    )
  }

  // Without history the API answers with popular issues and no "basedOn".
  const hasHistory = personal.some(item => item.basedOn)

  return (
    <main className="page">
      <h1 className="page-title">For you</h1>
      {(error || interactions.error) && <p className="error">{error || interactions.error}</p>}
      <section className="section">
        <h2 className="section-title">{hasHistory ? 'Based on your activity' : 'Popular issues'}</h2>
        <p className="muted">
          {hasHistory
            ? 'Issues that people with a history like yours worked on. Recent activity counts more.'
            : 'View or bookmark a few issues to get personal recommendations.'}
        </p>
        {rows(personal)}
      </section>
      {popular.length > 0 && (
        <section className="section">
          <h2 className="section-title">
            {languages.length > 0 ? 'Trending in ' + languages.join(', ') : 'Trending this week'}
          </h2>
          <p className="muted">
            Most active this week, your languages first. <Link to="/profile">Change languages</Link>
          </p>
          {rows(popular)}
        </section>
      )}
    </main>
  )
}
