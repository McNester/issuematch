import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { formatDate, splitByComma } from '../../shared/format.js'

const historyLabels = { view: 'Viewed', bookmark: 'Bookmarked', working: 'Working on it', solved: 'Solved' }
const historyFilters = ['all', 'bookmark', 'working', 'solved', 'view']

export default function ProfilePage() {
  const [form, setForm] = useState({ username: '', email: '', bio: '', languages: '', interests: '' })
  const [history, setHistory] = useState([])
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)
  const [filter, setFilter] = useState('all')

  useEffect(() => {
    api('/users/me')
      .then(user => setForm({
        username: user.username,
        email: user.email,
        bio: user.bio,
        languages: user.languages.join(', '),
        interests: user.interests.join(', '),
      }))
      .catch(failure => setError(failure.message))
    api('/users/me/interactions')
      .then(setHistory)
      .catch(failure => setError(failure.message))
  }, [])

  function field(name, label, type = 'text') {
    return (
      <div className="field">
        <label htmlFor={name}>{label}</label>
        <input id={name} type={type} className="input" value={form[name]}
          onChange={event => { setForm({ ...form, [name]: event.target.value }); setSaved(false) }} />
      </div>
    )
  }

  async function save(event) {
    event.preventDefault()
    try {
      await api('/users/me', {
        method: 'PUT',
        body: { ...form, languages: splitByComma(form.languages), interests: splitByComma(form.interests) },
      })
      setError('')
      setSaved(true)
    } catch (failure) {
      setError(failure.message)
    }
  }

  const shown = filter === 'all' ? history : history.filter(item => item.type === filter)

  return (
    <main className="page page-profile">
      <form className="profile-form" onSubmit={save}>
        <h1 className="page-title">Profile</h1>
        {field('username', 'Username')}
        {field('email', 'Email', 'email')}
        {field('bio', 'About')}
        {field('languages', 'Programming languages')}
        {field('interests', 'Interests')}
        {error && <p className="error">{error}</p>}
        {saved && <p className="muted">Changes saved.</p>}
        <button type="submit" className="button-primary button-start">Save changes</button>
      </form>
      <section className="profile-history">
        <h2 className="page-title">History</h2>
        <div className="filter-chips">
          {historyFilters.map(type => (
            <button key={type} type="button" className="filter-chip" aria-pressed={filter === type}
              onClick={() => setFilter(type)}>
              {type === 'all' ? 'All' : historyLabels[type]}
              <span className="filter-chip-count">
                {type === 'all' ? history.length : history.filter(item => item.type === type).length}
              </span>
            </button>
          ))}
        </div>
        {shown.length === 0 && <p className="muted">Nothing here yet.</p>}
        <div className="list">
          {shown.map(item => (
            <div key={item.id} className="history-row">
              <span className="history-type">{historyLabels[item.type]}</span>
              <Link to={'/issues/' + item.issueId} className="history-title">{item.title}</Link>
              <span className="history-date">{formatDate(item.createdAt)}</span>
            </div>
          ))}
        </div>
      </section>
    </main>
  )
}
