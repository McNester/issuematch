import { Link, useLocation } from 'react-router-dom'
import { endSession, isAdmin } from './session.js'

export default function Header() {
  const { pathname } = useLocation()
  const onIssues = pathname === '/' || pathname.startsWith('/issues')
  const linkClass = active => (active ? 'header-link active' : 'header-link')

  return (
    <header className="header">
      <Link to="/" className="header-brand">IssueMatch</Link>
      <nav className="header-nav">
        <Link to="/" className={linkClass(onIssues)}>Issues</Link>
        <Link to="/recommendations" className={linkClass(pathname === '/recommendations')}>For you</Link>
        <Link to="/profile" className={linkClass(pathname === '/profile')}>Profile</Link>
      </nav>
      {isAdmin() && <span className="chip">Admin</span>}
      <Link to="/signin" className="header-signout" onClick={endSession}>Sign out</Link>
    </header>
  )
}
