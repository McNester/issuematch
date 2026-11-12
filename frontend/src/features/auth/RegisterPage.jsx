import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { splitByComma } from '../../shared/format.js'
import { setToken } from '../../shared/session.js'

export default function RegisterPage() {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [languages, setLanguages] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  async function register(event) {
    event.preventDefault()
    try {
      const { token } = await api('/auth/register', {
        method: 'POST',
        body: { username, email, password, languages: splitByComma(languages) },
      })
      setToken(token)
      navigate('/')
    } catch (failure) {
      setError(failure.message)
    }
  }

  return (
    <div className="auth-page auth-page-register">
      <div className="auth-brand">IssueMatch</div>
      <main className="auth-card">
        <form className="auth-form" onSubmit={register}>
          <h1 className="auth-title">Create an account</h1>
          <div className="field">
            <label htmlFor="username">Username</label>
            <input id="username" type="text" className="input" placeholder="Username"
              value={username} onChange={event => setUsername(event.target.value)} />
          </div>
          <div className="field">
            <label htmlFor="email">Email</label>
            <input id="email" type="email" className="input" placeholder="you@example.com"
              value={email} onChange={event => setEmail(event.target.value)} />
          </div>
          <div className="field">
            <label htmlFor="password">Password</label>
            <input id="password" type="password" className="input" placeholder="At least 8 characters"
              value={password} onChange={event => setPassword(event.target.value)} />
          </div>
          <div className="field">
            <label htmlFor="languages">Programming languages</label>
            <input id="languages" type="text" className="input" placeholder="Java, Go"
              value={languages} onChange={event => setLanguages(event.target.value)} />
            <span className="field-hint">Separate with commas.</span>
          </div>
          {error && <p className="error">{error}</p>}
          <button type="submit" className="button-primary button-wide">Create account</button>
          <p className="auth-switch">Already registered? <Link to="/signin">Sign in</Link></p>
        </form>
      </main>
    </div>
  )
}
