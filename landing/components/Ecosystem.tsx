'use client'

import { useState } from 'react'

function StudioForm() {
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

  if (state === 'success') {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: 'var(--yellow)', fontSize: '14px', fontWeight: 500 }}>
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
          <path d="M3 8l3.5 3.5L13 4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        You&apos;re on the list.
      </div>
    )
  }

  return (
    <div>
      <form onSubmit={handleSubmit} style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
        <input
          type="email"
          required
          placeholder="you@company.com"
          value={email}
          onChange={e => setEmail(e.target.value)}
          disabled={state === 'loading'}
          style={{
            flex: '1 1 200px',
            background: 'rgba(255,255,255,0.07)',
            border: '1px solid rgba(255,255,255,0.15)',
            borderRadius: 'var(--radius-sm)',
            padding: '11px 14px',
            fontSize: '14px',
            color: '#fff',
            outline: 'none',
            fontFamily: 'var(--sans)',
          }}
        />
        <button
          type="submit"
          disabled={state === 'loading'}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '6px',
            background: 'var(--yellow)',
            color: 'var(--ink)',
            padding: '11px 18px',
            borderRadius: 'var(--radius-sm)',
            fontSize: '14px',
            fontWeight: 600,
            opacity: state === 'loading' ? 0.7 : 1,
            cursor: state === 'loading' ? 'default' : 'pointer',
            border: 'none',
            fontFamily: 'var(--sans)',
          }}
        >
          {state === 'loading' ? '…' : 'Request access'}
        </button>
      </form>
      {state === 'error' && (
        <p style={{ marginTop: '8px', fontSize: '13px', color: 'rgba(255,100,100,0.85)' }}>
          Something went wrong. Email <a href="mailto:daco@dacolabs.com" style={{ color: 'inherit', textDecoration: 'underline' }}>daco@dacolabs.com</a>.
        </p>
      )}
    </div>
  )
}

const features = [
  {
    title: 'Data product marketplace',
    body: 'Search and browse every data product your org has defined. One place, no Notion hunting.',
  },
  {
    title: 'Compliance & governance',
    body: 'Verify ownership, classify sensitivity, and enforce policy across your entire catalog automatically.',
  },
  {
    title: 'Quality monitoring',
    body: 'Quality rules from your OpenDPI specs become live monitors. Know when something drifts before downstream does.',
  },
  {
    title: 'Full lineage',
    body: 'Trace any field back to its source across every pipeline, team, and system. All derived from the specs your engineers already write.',
  },
]

export default function Ecosystem() {
  return (
    <section id="studio" style={{ background: 'var(--ink)', color: '#fff', padding: '96px 0', position: 'relative', overflow: 'hidden' }}>
      <div style={{
        position: 'absolute',
        inset: 0,
        backgroundImage: 'linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px)',
        backgroundSize: '64px 64px',
        maskImage: 'radial-gradient(ellipse 80% 60% at 50% 50%, #000 50%, transparent 100%)',
        pointerEvents: 'none',
      }} />
      <div className="container" style={{ position: 'relative', zIndex: 1 }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '72px', alignItems: 'start' }}>
          <div>
            <div className="eyebrow" style={{ color: 'rgba(255,255,255,0.45)' }}>
              <span className="dot" />Daco Studio · Early access
            </div>
            <h2 className="section-title" style={{ color: '#fff' }}>
              The cherry on top.
            </h2>
            <p className="section-lede" style={{ color: 'rgba(255,255,255,0.55)', marginBottom: '36px' }}>
              The CLI gives you the spec. Studio turns it into a live data catalog: search, governance, quality monitoring, and lineage, all driven by the OpenDPI definitions your engineers already maintain.
            </p>
            <StudioForm />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
            {features.map((f) => (
              <div
                key={f.title}
                style={{
                  background: 'rgba(255,255,255,0.04)',
                  border: '1px solid rgba(255,255,255,0.09)',
                  borderRadius: 'var(--radius)',
                  padding: '20px',
                }}
              >
                <div style={{ fontWeight: 600, fontSize: '14px', color: '#fff', marginBottom: '8px', lineHeight: 1.3 }}>
                  {f.title}
                </div>
                <div style={{ fontSize: '13.5px', color: 'rgba(255,255,255,0.5)', lineHeight: 1.55 }}>
                  {f.body}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
