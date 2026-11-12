import { useEffect, useState } from 'react'
import { api } from '../../shared/api.js'
import IssueRow from './IssueRow.jsx'

const pageSize = 20

export default function CatalogPage() {
  const [text, setText] = useState('')
  const [search, setSearch] = useState({ q: '', category: '', language: '', difficulty: '', page: 1 })
  const [result, setResult] = useState({ items: [], total: 0 })
  const [bookmarked, setBookmarked] = useState([])
  const [error, setError] = useState('')

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

  useEffect(() => {
    api('/users/me/interactions')
      .then(history => setBookmarked(history.filter(item => item.type === 'bookmark').map(item => item.issueId)))
      .catch(failure => setError(failure.message))
  }, [])

  function filterBy(name) {
    return event => setSearch({ ...search, [name]: event.target.value, page: 1 })
  }

  function submitSearch(event) {
    event.preventDefault()
    setSearch({ ...search, q: text, page: 1 })
  }

  async function bookmark(issueId) {
    try {
      await api('/interactions', { method: 'POST', body: { issueId, type: 'bookmark' } })
      setBookmarked([...bookmarked, issueId])
    } catch (failure) {
      setError(failure.message)
    }
  }

  const first = (search.page - 1) * pageSize + 1
  const last = first + result.items.length - 1

  return (
    <main className="page">
      <h1 className="page-title">Issues</h1>
      <form className="filters" onSubmit={submitSearch}>
        <input type="search" className="input filters-search" aria-label="Search issues by name"
          placeholder="Search by name or description" value={text} onChange={event => setText(event.target.value)} />
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
        <button type="submit" className="button-primary">Search</button>
      </form>
      {error && <p className="error">{error}</p>}
      <p className="muted">
        {result.total === 0
          ? 'No issues found'
          : `Showing ${first}–${last} of ${result.total.toLocaleString('en-US')} issues`}
      </p>
      <div className="list">
        {result.items.map(issue => (
          <IssueRow key={issue.id} issue={issue} bookmarked={bookmarked.includes(issue.id)} onBookmark={bookmark} />
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
