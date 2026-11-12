import { useEffect, useState } from 'react'
import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { capitalize, formatDate } from '../../shared/format.js'
import { isAdmin } from '../../shared/session.js'
import { IssueTags } from './IssueRow.jsx'
import { useInteractions } from './useInteractions.js'

const actions = [
  { type: 'bookmark', off: 'Bookmark', on: 'Bookmarked' },
  { type: 'working', off: 'Working on it', on: 'Working on it' },
  { type: 'solved', off: 'Mark as solved', on: 'Solved' },
]

// An imported issue has an externalId such as "github:golang/go#812".
function githubUrl(externalId) {
  const [, repository, number] = externalId.match(/^github:(.+)#(\d+)$/) || []
  return repository && `https://github.com/${repository}/issues/${number}`
}

// Issue templates on GitHub leave HTML comments in the text.
function withoutComments(markdown) {
  return markdown.replace(/<!--[\s\S]*?-->/g, '')
}

export default function IssuePage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [issue, setIssue] = useState(null)
  const [error, setError] = useState('')
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const interactions = useInteractions()

  useEffect(() => {
    api('/issues/' + id)
      .then(setIssue)
      .then(() => api('/interactions', { method: 'POST', body: { issueId: id, type: 'view' } }))
      .catch(failure => setError(failure.message))
  }, [id])

  async function deleteIssue() {
    try {
      await api('/issues/' + id, { method: 'DELETE' })
      navigate('/')
    } catch (failure) {
      setError(failure.message)
    }
  }

  const source = issue?.externalId && githubUrl(issue.externalId)

  return (
    <main className="page page-issue">
      <Link to="/" className="back-link">Back to issues</Link>
      {(error || interactions.error) && <p className="error">{error || interactions.error}</p>}
      {issue && (
        <div className="issue-layout">
          <article className="issue-main">
            <h1 className="issue-title">{issue.title}</h1>
            <IssueTags issue={issue} />
            <div className="issue-actions">
              {actions.map(action => {
                const on = interactions.has(id, action.type)
                return (
                  <button key={action.type} type="button" aria-pressed={on}
                    className={action.type === 'solved' && !on ? 'button-primary button-compact' : 'button-secondary button-strong'}
                    onClick={() => interactions.toggle(id, action.type)}>
                    {on && <span aria-hidden="true">✓ </span>}{on ? action.on : action.off}
                  </button>
                )
              })}
              {source && (
                <a href={source} target="_blank" rel="noreferrer" className="button-secondary button-strong button-link">
                  View on GitHub
                </a>
              )}
            </div>
            <div className="markdown">
              <Markdown remarkPlugins={[remarkGfm]}
                components={{ a: props => <a {...props} target="_blank" rel="noreferrer" /> }}>
                {withoutComments(issue.description)}
              </Markdown>
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
            {isAdmin() && (
              <div className="details-admin">
                <Link to={`/issues/${id}/edit`} className="button-secondary button-link">Edit</Link>
                {confirmingDelete
                  ? <button type="button" className="button-danger" onClick={deleteIssue}>Confirm delete</button>
                  : <button type="button" className="button-secondary" onClick={() => setConfirmingDelete(true)}>Delete</button>}
              </div>
            )}
          </aside>
        </div>
      )}
    </main>
  )
}
