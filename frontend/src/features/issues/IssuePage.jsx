import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { capitalize, formatDate } from '../../shared/format.js'
import { IssueTags } from './IssueRow.jsx'

export default function IssuePage() {
  const { id } = useParams()
  const [issue, setIssue] = useState(null)
  const [recorded, setRecorded] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function open() {
      try {
        setIssue(await api('/issues/' + id))
        await api('/interactions', { method: 'POST', body: { issueId: id, type: 'view' } })
        const history = await api('/users/me/interactions')
        setRecorded(history.filter(item => item.issueId === id).map(item => item.type))
      } catch (failure) {
        setError(failure.message)
      }
    }
    open()
  }, [id])

  async function record(type) {
    try {
      await api('/interactions', { method: 'POST', body: { issueId: id, type } })
      setRecorded([...recorded, type])
    } catch (failure) {
      setError(failure.message)
    }
  }

  return (
    <main className="page page-issue">
      <Link to="/" className="back-link">Back to issues</Link>
      {error && <p className="error">{error}</p>}
      {issue && (
        <div className="issue-layout">
          <article className="issue-main">
            <h1 className="issue-title">{issue.title}</h1>
            <IssueTags issue={issue} />
            <div className="issue-description">
              {issue.description.split(/\n\s*\n/).map((paragraph, index) => <p key={index}>{paragraph}</p>)}
            </div>
            <div className="issue-actions">
              <button type="button" className="button-secondary button-strong"
                aria-pressed={recorded.includes('bookmark')} onClick={() => record('bookmark')}>Bookmark</button>
              <button type="button" className="button-secondary button-strong"
                aria-pressed={recorded.includes('working')} onClick={() => record('working')}>Working on it</button>
              <button type="button" className="button-primary button-compact"
                aria-pressed={recorded.includes('solved')} onClick={() => record('solved')}>Mark as solved</button>
            </div>
            <p className="issue-view-note">Opening this page is recorded as a view.</p>
          </article>
          <aside className="details">
            <h2 className="details-title">Details</h2>
            <div className="details-row"><span>Project</span><span className="repo repo-dark">{issue.project.name}</span></div>
            <div className="details-row"><span>Category</span><span>{capitalize(issue.category)}</span></div>
            <div className="details-row"><span>Language</span><span>{issue.language}</span></div>
            <div className="details-row"><span>Difficulty</span><span>{capitalize(issue.difficulty)}</span></div>
            <div className="details-row"><span>Status</span><span>{capitalize(issue.status)}</span></div>
            <div className="details-row"><span>Created</span><span>{formatDate(issue.createdAt)}</span></div>
          </aside>
        </div>
      )}
    </main>
  )
}
