import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { setToken } from '../../shared/session.js'

export default function SignInPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  async function signIn(event) {
    event.preventDefault()
    try {
      const { token } = await api('/auth/login', { method: 'POST', body: { email, password } })
      setToken(token)
      navigate('/')
    } catch (failure) {
      setError(failure.message)
    }
  }

  return (
    <div className="auth-page auth-page-signin">
      <div className="auth-brand">IssueMatch</div>
      <main className="auth-card">
        <form className="auth-form" onSubmit={signIn}>
          <h1 className="auth-title">Sign in</h1>
          <div className="field">
            <label htmlFor="email">Email</label>
            <input id="email" type="email" className="input" placeholder="you@example.com"
              value={email} onChange={event => setEmail(event.target.value)} />
          </div>
          <div className="field">
            <label htmlFor="password">Password</label>
            <input id="password" type="password" className="input" placeholder="Password"
              value={password} onChange={event => setPassword(event.target.value)} />
          </div>
          {error && <p className="error">{error}</p>}
          <button type="submit" className="button-primary button-wide">Sign in</button>
          <p className="auth-switch">No account yet? <Link to="/register">Create an account</Link></p>
        </form>
      </main>
    </div>
  )
}
