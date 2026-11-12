import { Link, useLocation } from 'react-router-dom'
import { clearToken } from './session.js'

export default function Header() {
  const { pathname } = useLocation()
  const onIssues = pathname === '/' || pathname.startsWith('/issues')

  return (
    <header className="header">
      <Link to="/" className="header-brand">IssueMatch</Link>
      <nav className="header-nav">
        <Link to="/" className={onIssues ? 'header-link active' : 'header-link'}>Issues</Link>
        <Link to="/recommendations" className={pathname === '/recommendations' ? 'header-link active' : 'header-link'}>For you</Link>
        <Link to="/profile" className={pathname === '/profile' ? 'header-link active' : 'header-link'}>Profile</Link>
      </nav>
      <Link to="/signin" className="header-signout" onClick={clearToken}>Sign out</Link>
    </header>
  )
}
