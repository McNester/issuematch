import { Link } from 'react-router-dom'
import { capitalize } from '../../shared/format.js'

export default function IssueRow({ issue, bookmarked, onBookmark, note }) {
  return (
    <div className="issue-row">
      <div className="issue-row-body">
        <Link to={'/issues/' + issue.id} className="issue-row-title">{issue.title}</Link>
        <IssueTags issue={issue} showProject />
        {note && <span className="issue-row-note">{note}</span>}
      </div>
      <button type="button" className="button-icon" aria-pressed={bookmarked}
        aria-label={bookmarked ? 'Remove the bookmark' : 'Bookmark this issue'}
        title={bookmarked ? 'Bookmarked' : 'Bookmark'} onClick={() => onBookmark(issue.id)}>
        <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path>
        </svg>
      </button>
    </div>
  )
}

export function IssueTags({ issue, showProject }) {
  return (
    <div className="tags">
      {showProject && <span className="repo">{issue.project.name}</span>}
      <span className="chip">{capitalize(issue.category)}</span>
      {issue.language && <span className="chip">{issue.language}</span>}
      <span className="chip">{capitalize(issue.difficulty)}</span>
      {issue.labels.map(label => <span key={label} className="label">{label}</span>)}
    </div>
  )
}
