import { Navigate, Outlet, Route, Routes } from 'react-router-dom'
import Header from './shared/Header.jsx'
import { getToken } from './shared/session.js'
import SignInPage from './features/auth/SignInPage.jsx'
import RegisterPage from './features/auth/RegisterPage.jsx'
import CatalogPage from './features/issues/CatalogPage.jsx'
import IssuePage from './features/issues/IssuePage.jsx'
import ProfilePage from './features/profile/ProfilePage.jsx'
import RecommendationsPage from './features/recommendations/RecommendationsPage.jsx'

function SignedIn() {
  if (!getToken()) {
    return <Navigate to="/signin" replace />
  }
  return (
    <>
      <Header />
      <Outlet />
    </>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/signin" element={<SignInPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route element={<SignedIn />}>
        <Route path="/" element={<CatalogPage />} />
        <Route path="/issues/:id" element={<IssuePage />} />
        <Route path="/recommendations" element={<RecommendationsPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Route>
    </Routes>
  )
}
