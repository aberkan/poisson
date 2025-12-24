import React, { useState, useEffect } from 'react'
import { request } from 'graphql-request'
import '../App.css'

const GRAPHQL_ENDPOINT = 'http://localhost:8080/graphql'

const FEED_QUERY = `
  query Feed($maxArticles: Int!, $oldestDate: String!, $mode: String!) {
    feed(maxArticles: $maxArticles, oldestDate: $oldestDate, mode: $mode) {
      url
      title
      jokeConfidence
    }
  }
`

function FeedPage() {
  const [feedItems, setFeedItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function fetchFeed() {
      try {
        setLoading(true)
        setError(null)

        // Calculate date 1 week ago in YYYY-MM-DD format
        const oneWeekAgo = new Date()
        oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)
        const oldestDate = oneWeekAgo.toISOString().split('T')[0] // Format as YYYY-MM-DD

        const variables = {
          maxArticles: 10,
          oldestDate: oldestDate,
          mode: 'joke'
        }

        const data = await request(GRAPHQL_ENDPOINT, FEED_QUERY, variables)
        setFeedItems(data.feed || [])
      } catch (err) {
        setError(err.message)
        setFeedItems([])
      } finally {
        setLoading(false)
      }
    }

    fetchFeed()
  }, [])

  return (
    <div className="App">
      <h1>Feed - Top Joke Articles</h1>
      <p className="info">Showing articles from the past week, ranked by joke confidence</p>

      {loading && <p>Loading feed...</p>}
      
      {error && <p className="error">Error: {error}</p>}

      {!loading && !error && (
        <>
          {feedItems.length === 0 ? (
            <p className="info">No articles found in the feed.</p>
          ) : (
            <div className="feed-section">
              <h2>Articles ({feedItems.length})</h2>
              <div className="feed-list">
                {feedItems.map((item, index) => (
                  <div key={index} className="feed-item">
                    <div className="feed-item-header">
                      <span className="feed-rank">#{index + 1}</span>
                      <span className="feed-confidence">{item.jokeConfidence}%</span>
                    </div>
                    <h3 className="feed-title">
                      <a href={`/analysis?url=${encodeURIComponent(item.url)}`}>
                        {item.title}
                      </a>
                    </h3>
                    <p className="feed-url">
                      <a href={item.url} target="_blank" rel="noopener noreferrer">
                        {item.url}
                      </a>
                    </p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default FeedPage

