export default function FinalCTA() {
  return (
    <section className="final-cta" id="demo">
      <div className="container final-cta-inner">
        <h2>Start with <em>your</em> repo.</h2>
        <p>Connect Daco Studio to your repository and get a live catalog of your data products in minutes.</p>
        <div style={{ display: 'flex', gap: '16px', justifyContent: 'center', flexWrap: 'wrap' }}>
          <a href="/studio" className="btn-primary">
            Try Daco Studio
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M2 7h10m0 0L7 2m5 5l-5 5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </a>
          <a href="/docs" style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', fontSize: '15px', fontWeight: 600, color: 'rgba(255,255,255,0.7)', padding: '13px 22px', border: '1px solid rgba(255,255,255,0.15)', borderRadius: 'var(--radius-sm)', transition: 'color 0.15s, border-color 0.15s' }}>
            Read the docs
          </a>
        </div>
      </div>
    </section>
  )
}
