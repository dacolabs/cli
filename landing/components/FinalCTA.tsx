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
        <h2>Get <em>early</em> access.</h2>
        <p>We&apos;re onboarding teams one by one. Drop your email and we&apos;ll reach out.</p>

        {state === 'success' ? (
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '10px', background: 'rgba(244,207,74,0.1)', border: '1px solid rgba(244,207,74,0.3)', borderRadius: 'var(--radius-sm)', padding: '14px 24px', fontSize: '15px', color: 'var(--yellow)', fontWeight: 500 }}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path d="M3 8l3.5 3.5L13 4.5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            You&apos;re on the list. We&apos;ll be in touch.
          </div>
        ) : (
          <form onSubmit={handleSubmit} style={{ display: 'flex', gap: '10px', justifyContent: 'center', flexWrap: 'wrap', maxWidth: '480px', margin: '0 auto' }}>
            <input
              type="email"
              required
              placeholder="you@company.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              disabled={state === 'loading'}
              style={{
                flex: 1,
                minWidth: '220px',
                background: 'rgba(255,255,255,0.06)',
                border: '1px solid rgba(255,255,255,0.15)',
                borderRadius: 'var(--radius-sm)',
                padding: '13px 16px',
                fontSize: '15px',
                color: '#fff',
                fontFamily: 'var(--sans)',
                outline: 'none',
              }}
            />
            <button
              type="submit"
              disabled={state === 'loading'}
              className="btn-primary"
              style={{ opacity: state === 'loading' ? 0.7 : 1 }}
            >
              {state === 'loading' ? 'Sending…' : 'Get early access'}
            </button>
          </form>
        )}

        {state === 'error' && (
          <p style={{ marginTop: '12px', fontSize: '14px', color: 'rgba(255,100,100,0.9)' }}>
            Something went wrong. Email us directly at{' '}
            <a href="mailto:daco@dacolabs.com" style={{ color: 'var(--yellow)' }}>daco@dacolabs.com</a>.
          </p>
        )}
      </div>
    </section>
  )
}
