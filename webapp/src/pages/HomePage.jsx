import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { request } from 'graphql-request'
import '../App.css'

const GRAPHQL_ENDPOINT = 'http://localhost:8080/graphql'

const HEALTH_QUERY = `
  query {
    health
  }
`

function HomePage() {
  const [healthStatus, setHealthStatus] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function fetchHealth() {
      try {
        setLoading(true)
        const data = await request(GRAPHQL_ENDPOINT, HEALTH_QUERY)
        setHealthStatus(data.health)
        setError(null)
      } catch (err) {
        setError(err.message)
        setHealthStatus(null)
      } finally {
        setLoading(false)
      }
    }

    fetchHealth()
  }, [])

  return (
    <div className="App">
      <h1>Welcome to Poisson</h1>
      <div className="health-status">
        {loading && <p>Checking health status...</p>}
        {error && <p className="error">Error: {error}</p>}
        {healthStatus && (
          <p className="health-result">Health Status: {healthStatus}</p>
        )}
      </div>
      <div className="home-links">
        <Link to="/feed" className="feed-link">Go see the feed</Link>
      </div>
    </div>
  )
}

export default HomePage

