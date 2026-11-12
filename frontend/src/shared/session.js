const tokenKey = 'issuematch.token'
const roleKey = 'issuematch.role'

export function getToken() {
  return localStorage.getItem(tokenKey)
}

export function isAdmin() {
  return localStorage.getItem(roleKey) === 'admin'
}

export function startSession({ token, user }) {
  localStorage.setItem(tokenKey, token)
  localStorage.setItem(roleKey, user.role)
}

export function endSession() {
  localStorage.removeItem(tokenKey)
  localStorage.removeItem(roleKey)
}
