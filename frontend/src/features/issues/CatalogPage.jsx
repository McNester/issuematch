import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { isAdmin } from '../../shared/session.js'
import RecomputeButton from '../recommendations/RecomputeButton.jsx'
import IssueRow from './IssueRow.jsx'
import Trending from './Trending.jsx'
import { useInteractions } from './useInteractions.js'

const pageSize = 20

export default function CatalogPage() {
  const [text, setText] = useState('')
  const [search, setSearch] = useState({ q: '', category: '', language: '', difficulty: '', page: 1 })
  const [result, setResult] = useState({ items: [], total: 0 })
  const [error, setError] = useState('')
  const interactions = useInteractions()

  useEffect(() => {
    const params = new URLSearchParams({ page: search.page, limit: pageSize })
    for (const name of ['q', 'category', 'language', 'difficulty']) {
      if (search[name]) {
        params.set(name, search[name])
      }
    }
    api('/issues?' + params)
      .then(found => { setResult(found); setError('') })
      .catch(failure => setError(failure.message))
  }, [search])

  function filterBy(name) {
    return event => setSearch({ ...search, [name]: event.target.value, page: 1 })
  }

  function submitSearch(event) {
    event.preventDefault()
    setSearch({ ...search, q: text, page: 1 })
  }

  const browsing = !search.q && !search.category && !search.language && !search.difficulty && search.page === 1

  const first = (search.page - 1) * pageSize + 1
  const last = first + result.items.length - 1

  return (
    <main className="page">
      <div className="page-heading">
        <h1 className="page-title">Issues</h1>
        {isAdmin() && (
          <div className="page-heading-actions">
            <RecomputeButton />
            <Link to="/issues/new" className="button-secondary button-link">New issue</Link>
          </div>
        )}
      </div>
      <form className="filters" onSubmit={submitSearch}>
        <div className="filters-search">
          <input type="search" className="input" aria-label="Search issues by name"
            placeholder="Search by name or description" value={text} onChange={event => setText(event.target.value)} />
          <button type="submit" className="button-primary">Search</button>
        </div>
        <select className="select" aria-label="Category" value={search.category} onChange={filterBy('category')}>
          <option value="">All categories</option>
          <option value="bug">Bug</option>
          <option value="feature">Feature</option>
          <option value="documentation">Documentation</option>
          <option value="refactoring">Refactoring</option>
        </select>
        <select className="select" aria-label="Language" value={search.language} onChange={filterBy('language')}>
          <option value="">All languages</option>
          <option>Go</option>
          <option>Java</option>
          <option>JavaScript</option>
          <option>Python</option>
          <option>Rust</option>
        </select>
        <select className="select" aria-label="Difficulty" value={search.difficulty} onChange={filterBy('difficulty')}>
          <option value="">Any difficulty</option>
          <option value="easy">Easy</option>
          <option value="medium">Medium</option>
          <option value="hard">Hard</option>
        </select>
      </form>
      {(error || interactions.error) && <p className="error">{error || interactions.error}</p>}
      {browsing && <Trending />}
      <p className="muted">
        {result.total === 0
          ? 'No issues found'
          : `Showing ${first}–${last} of ${result.total.toLocaleString('en-US')} issues`}
      </p>
      <div className="list">
        {result.items.map(issue => (
          <IssueRow key={issue.id} issue={issue} bookmarked={interactions.has(issue.id, 'bookmark')}
            onBookmark={issueId => interactions.toggle(issueId, 'bookmark')} />
        ))}
      </div>
      <div className="pagination">
        <button type="button" className="button-secondary" disabled={search.page === 1}
          onClick={() => setSearch({ ...search, page: search.page - 1 })}>Previous</button>
        <span className="muted">Page {search.page}</span>
        <button type="button" className="button-secondary" disabled={last >= result.total}
          onClick={() => setSearch({ ...search, page: search.page + 1 })}>Next</button>
      </div>
    </main>
  )
}
