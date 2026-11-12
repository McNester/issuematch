import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../shared/api.js'
import { IssueTags } from './IssueRow.jsx'

export default function Trending() {
  const [items, setItems] = useState([])

  useEffect(() => {
    // The catalog works without this strip, so a failure here stays silent.
    api('/recommendations/trending?limit=3').then(response => setItems(response.items)).catch(() => {})
  }, [])

  if (items.length === 0) {
    return null
  }
  return (
    <section className="section">
      <h2 className="section-title">Trending this week</h2>
      <div className="cards">
        {items.map(({ issue }) => (
          <Link key={issue.id} to={'/issues/' + issue.id} className="card">
            <span className="repo">{issue.project.name}</span>
            <span className="card-title">{issue.title}</span>
            <IssueTags issue={{ ...issue, labels: [] }} />
          </Link>
        ))}
      </div>
    </section>
  )
}
