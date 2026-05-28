'use client'

import { useState } from 'react'

export default function FinalCTA() {
  const [email, setEmail] = useState('')
  const [state, setState] = useState<'idle' | 'loading' | 'success' | 'error'>('idle')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!email) return
    setState('loading')
    try {
      const res = await fetch('/api/waitlist', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })
      if (!res.ok) throw new Error()
      setState('success')
      setEmail('')
    } catch {
      setState('error')
    }
  }

  return (
    <section className="final-cta" id="demo">
      <div className="container final-cta-inner">
        <h2>Pick your <em>on-ramp</em>.</h2>
        <p>The CLI is a one-liner away. The platform is in early access.</p>

        <div className="final-cta-grid">
          <div className="final-cta-card">
            <div className="final-cta-card-eyebrow">Install</div>
            <div className="final-cta-card-title">Start with the CLI</div>
            <div className="final-cta-card-body">Free and open source. Brew, then <code>daco init</code>.</div>
            <a href="/docs" className="final-cta-card-link">
              Read the docs
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                <path d="M2 6h8m0 0L6 2m4 4L6 10" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </a>
          </div>

          <div className="final-cta-card highlighted">
            <div className="final-cta-card-eyebrow">Platform</div>
            <div className="final-cta-card-title">Get Studio access</div>
            <div className="final-cta-card-body">We onboard teams one by one. Drop your email and we&apos;ll reach out.</div>
            {state === 'success' ? (
              <div className="final-cta-success">
                <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                  <path d="M3 8l3.5 3.5L13 4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                You&apos;re on the list.
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="final-cta-form">
                <input
                  type="email"
                  required
                  placeholder="you@company.com"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  disabled={state === 'loading'}
                />
                <button
                  type="submit"
                  disabled={state === 'loading'}
                  className="btn-primary"
                  style={{ opacity: state === 'loading' ? 0.7 : 1 }}
                >
                  {state === 'loading' ? '…' : 'Request Access'}
                </button>
              </form>
            )}
            {state === 'error' && (
              <p className="final-cta-error">
                Something went wrong. Email{' '}
                <a href="mailto:daco@dacolabs.com">daco@dacolabs.com</a>.
              </p>
            )}
          </div>

        </div>
      </div>
    </section>
  )
}
