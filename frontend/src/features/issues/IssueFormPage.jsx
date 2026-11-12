import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { splitByComma } from '../../shared/format.js'

const emptyForm = {
  title: '', description: '', category: 'bug', language: '', difficulty: 'medium', status: 'open',
  labels: '', projectName: '', projectUrl: '', projectStars: '0',
}

// Admins create an issue at /issues/new and edit one at /issues/:id/edit.
export default function IssueFormPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState(emptyForm)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) {
      return
    }
    api('/issues/' + id)
      .then(issue => setForm({
        title: issue.title,
        description: issue.description,
        category: issue.category,
        language: issue.language,
        difficulty: issue.difficulty,
        status: issue.status,
        labels: issue.labels.join(', '),
        projectName: issue.project.name,
        projectUrl: issue.project.url,
        projectStars: String(issue.project.stars),
      }))
      .catch(failure => setError(failure.message))
  }, [id])

  function control(name) {
    return { id: name, value: form[name], onChange: event => setForm({ ...form, [name]: event.target.value }) }
  }

  async function save(event) {
    event.preventDefault()
    const body = {
      title: form.title,
      description: form.description,
      category: form.category,
      language: form.language,
      difficulty: form.difficulty,
      status: form.status,
      labels: splitByComma(form.labels),
      project: { name: form.projectName, url: form.projectUrl, stars: Number(form.projectStars) || 0 },
    }
    try {
      const saved = await api(id ? '/issues/' + id : '/issues', { method: id ? 'PUT' : 'POST', body })
      navigate('/issues/' + saved.id)
    } catch (failure) {
      setError(failure.message)
    }
  }

  return (
    <main className="page page-form">
      <h1 className="page-title">{id ? 'Edit issue' : 'New issue'}</h1>
      <form className="form" onSubmit={save}>
        <div className="field">
          <label htmlFor="title">Title</label>
          <input type="text" className="input" {...control('title')} />
        </div>
        <div className="field">
          <label htmlFor="description">Description</label>
          <textarea className="input textarea" rows="8" {...control('description')} />
          <span className="field-hint">Markdown is supported.</span>
        </div>
        <div className="form-row">
          <div className="field">
            <label htmlFor="category">Category</label>
            <select className="select" {...control('category')}>
              <option value="bug">Bug</option>
              <option value="feature">Feature</option>
              <option value="documentation">Documentation</option>
              <option value="refactoring">Refactoring</option>
            </select>
          </div>
          <div className="field">
            <label htmlFor="difficulty">Difficulty</label>
            <select className="select" {...control('difficulty')}>
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
            </select>
          </div>
          <div className="field">
            <label htmlFor="status">Status</label>
            <select className="select" {...control('status')}>
              <option value="open">Open</option>
              <option value="closed">Closed</option>
            </select>
          </div>
        </div>
        <div className="form-row">
          <div className="field">
            <label htmlFor="language">Language</label>
            <input type="text" className="input" placeholder="Go" {...control('language')} />
          </div>
          <div className="field">
            <label htmlFor="labels">Labels</label>
            <input type="text" className="input" placeholder="good first issue, help wanted" {...control('labels')} />
          </div>
        </div>
        <div className="form-row">
          <div className="field">
            <label htmlFor="projectName">Project</label>
            <input type="text" className="input" placeholder="owner/name" {...control('projectName')} />
          </div>
          <div className="field">
            <label htmlFor="projectUrl">Project URL</label>
            <input type="url" className="input" placeholder="https://github.com/owner/name" {...control('projectUrl')} />
          </div>
          <div className="field">
            <label htmlFor="projectStars">Stars</label>
            <input type="number" min="0" className="input" {...control('projectStars')} />
          </div>
        </div>
        {error && <p className="error">{error}</p>}
        <div className="form-actions">
          <button type="submit" className="button-primary">{id ? 'Save changes' : 'Create issue'}</button>
          <Link to={id ? '/issues/' + id : '/'} className="button-secondary button-link">Cancel</Link>
        </div>
      </form>
    </main>
  )
}
