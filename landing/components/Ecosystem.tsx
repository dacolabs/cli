import Link from 'next/link'

export default function Ecosystem() {
  return (
    <section id="ecosystem" className="studio-banner-section">
      <div className="container">
        <Link href="/studio" className="studio-banner">
          <div className="studio-banner-glyph">
            <svg viewBox="0 0 80 80" width="56" height="56" fill="none" stroke="currentColor" strokeWidth="1.8">
              <rect x="6" y="6" width="32" height="32" rx="3"/>
              <rect x="42" y="6" width="32" height="32" rx="3"/>
              <rect x="6" y="42" width="32" height="32" rx="3"/>
              <rect x="42" y="42" width="32" height="32" rx="3" fill="currentColor"/>
            </svg>
          </div>
          <div className="studio-banner-body">
            <div className="studio-banner-tag">Daco Studio · Early access</div>
            <h3>There&apos;s also a control room for governance teams.</h3>
            <p>Search your data product marketplace, verify compliance, monitor quality rules, and trace ownership across every dataset — built on the same OpenDPI specs your engineers already write.</p>
          </div>
          <div className="studio-banner-arrow">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <path d="M3 10h14m0 0L11 4m6 6l-6 6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
        </Link>
      </div>
    </section>
  )
}
