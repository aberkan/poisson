import React, { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { request } from 'graphql-request'
import '../App.css'

const GRAPHQL_ENDPOINT = 'http://localhost:8080/graphql'

const ANALYSIS_QUERY = `
  query Analysis($url: String!, $mode: String) {
    analysis(url: $url, mode: $mode) {
      mode
      jokePercentage
      jokeReasoning
      promptFingerprint
    }
  }
`

const CRAWLED_PAGE_QUERY = `
  query CrawledPage($url: String!) {
    crawledPage(url: $url) {
      url
      title
      content
      datetime
    }
  }
`

function AnalysisPage() {
  const [searchParams] = useSearchParams()
  const url = searchParams.get('url')
  
  const [analysis, setAnalysis] = useState(null)
  const [crawledPage, setCrawledPage] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function fetchData() {
      if (!url) {
        setError('No URL provided in query parameters')
        setLoading(false)
        return
      }

      try {
        setLoading(true)
        setError(null)

        // Fetch both analysis and crawled page in parallel
        const [analysisData, crawledPageData] = await Promise.all([
          request(GRAPHQL_ENDPOINT, ANALYSIS_QUERY, { url }),
          request(GRAPHQL_ENDPOINT, CRAWLED_PAGE_QUERY, { url }).catch(() => null)
        ])

        setAnalysis(analysisData.analysis)
        setCrawledPage(crawledPageData?.crawledPage || null)
      } catch (err) {
        setError(err.message)
        setAnalysis(null)
        setCrawledPage(null)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [url])

  if (!url) {
    return (
      <div className="App">
        <h1>Analysis Page</h1>
        <p className="error">No URL provided. Please add ?url=YOUR_URL to the query string.</p>
      </div>
    )
  }

  return (
    <div className="App">
      <h1>Analysis for URL</h1>
      <p className="url-display">URL: {url}</p>

      {loading && <p>Loading...</p>}
      
      {error && <p className="error">Error: {error}</p>}

      {!loading && !error && (
        <>
          {crawledPage ? (
            <div className="crawled-page-section">
              <h2>Crawled Page</h2>
              <div className="info-box">
                <p><strong>Title:</strong> {crawledPage.title}</p>
                <p><strong>Crawled At:</strong> {crawledPage.datetime}</p>
                <p><strong>Content Length:</strong> {crawledPage.content.length} characters</p>
                <details>
                  <summary>View Content</summary>
                  <pre className="content-preview">{crawledPage.content}</pre>
                </details>
              </div>
            </div>
          ) : (
            <p className="info">No crawled page found for this URL.</p>
          )}

          {analysis ? (
            <div className="analysis-section">
              <h2>Analysis Results</h2>
              <div className="info-box">
                <p><strong>Mode:</strong> {analysis.mode}</p>
                {analysis.jokePercentage !== null && analysis.jokePercentage !== undefined && (
                  <p><strong>Joke Percentage:</strong> {analysis.jokePercentage}%</p>
                )}
                {analysis.jokeReasoning && (
                  <div>
                    <p><strong>Joke Reasoning:</strong></p>
                    <p className="reasoning-text">{analysis.jokeReasoning}</p>
                  </div>
                )}
                <p><strong>Prompt Fingerprint:</strong> {analysis.promptFingerprint}</p>
              </div>
            </div>
          ) : (
            <p className="info">No analysis found for this URL.</p>
          )}
        </>
      )}
    </div>
  )
}

export default AnalysisPage

