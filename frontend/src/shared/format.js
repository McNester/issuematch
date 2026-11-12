export function capitalize(text) {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

export function formatDate(isoDate) {
  return new Date(isoDate).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

export function splitByComma(text) {
  return text.split(',').map(part => part.trim()).filter(Boolean)
}
