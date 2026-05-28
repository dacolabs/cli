import Link from 'next/link'
import Logo from './Logo'

export default function Nav() {
  return (
    <nav className="nav">
      <div className="container nav-inner">
        <Link href="/" className="logo" style={{ color: '#fff' }}>
          <Logo size={26} />
          daco
        </Link>
        <div className="nav-right">
          <Link href="/docs" style={{ fontSize: '14px', fontWeight: 500, color: 'rgba(255,255,255,0.7)' }}>Docs</Link>
          <Link href="/blog" style={{ fontSize: '14px', fontWeight: 500, color: 'rgba(255,255,255,0.7)' }}>Blog</Link>
          <a href="https://github.com/dacolabs/cli" target="_blank" rel="noopener noreferrer" style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', color: 'rgba(255,255,255,0.7)', fontSize: '13.5px', fontWeight: 500, padding: '6px 10px', border: '1px solid rgba(255,255,255,0.12)', borderRadius: '6px', transition: 'color 0.15s, border-color 0.15s' }}>
            <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 0C3.58 0 0 3.58 0 8a8 8 0 0 0 5.47 7.59c.4.07.55-.17.55-.38v-1.33c-2.22.48-2.69-1.07-2.69-1.07-.36-.92-.89-1.17-.89-1.17-.73-.5.05-.49.05-.49.81.06 1.23.83 1.23.83.72 1.23 1.88.88 2.34.67.07-.52.28-.88.51-1.08-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.13 0 0 .67-.21 2.2.82a7.65 7.65 0 0 1 4 0c1.53-1.04 2.2-.82 2.2-.82.44 1.11.16 1.93.08 2.13.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48v2.2c0 .21.15.46.55.38A8 8 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            GitHub
          </a>
          <a href="#demo" className="nav-cta">
            Get started
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
              <path d="M2 5.5h7m0 0L5.5 2M9 5.5L5.5 9" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </a>
        </div>
      </div>
    </nav>
  )
}
