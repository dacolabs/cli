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
        <div className="nav-links">
          <a href="/docs">Docs</a>
          <a href="/studio">Studio</a>
          <a href="/blog">Blog</a>
        </div>
        <div className="nav-right">
          <a href="https://discord.gg/dacolabs" target="_blank" rel="noopener noreferrer" style={{ color: 'rgba(255,255,255,0.7)' }} aria-label="Discord">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M13.55 2.95A12.7 12.7 0 0 0 10.4 2c-.14.25-.3.59-.42.86a11.7 11.7 0 0 0-3.95 0A8.5 8.5 0 0 0 5.6 2a12.6 12.6 0 0 0-3.16.95C.43 5.96-.12 8.9.16 11.79a12.8 12.8 0 0 0 3.87 1.96c.31-.43.59-.88.83-1.36-.46-.17-.9-.39-1.32-.64.11-.08.22-.17.32-.25a9.13 9.13 0 0 0 7.88 0c.1.09.21.17.32.25-.42.25-.86.47-1.32.64.24.48.52.93.83 1.36a12.8 12.8 0 0 0 3.87-1.96c.34-3.36-.55-6.27-2.4-8.84zM5.34 9.96c-.76 0-1.39-.7-1.39-1.56s.62-1.56 1.39-1.56 1.4.7 1.39 1.56c0 .86-.62 1.56-1.39 1.56zm5.32 0c-.76 0-1.39-.7-1.39-1.56s.62-1.56 1.39-1.56 1.4.7 1.39 1.56c0 .86-.62 1.56-1.39 1.56z"/>
            </svg>
          </a>
          <a href="https://github.com/dacolabs/daco" target="_blank" rel="noopener noreferrer" style={{ color: 'rgba(255,255,255,0.7)' }} aria-label="GitHub">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 0C3.58 0 0 3.58 0 8a8 8 0 0 0 5.47 7.59c.4.07.55-.17.55-.38v-1.33c-2.22.48-2.69-1.07-2.69-1.07-.36-.92-.89-1.17-.89-1.17-.73-.5.05-.49.05-.49.81.06 1.23.83 1.23.83.72 1.23 1.88.88 2.34.67.07-.52.28-.88.51-1.08-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.13 0 0 .67-.21 2.2.82a7.65 7.65 0 0 1 4 0c1.53-1.04 2.2-.82 2.2-.82.44 1.11.16 1.93.08 2.13.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48v2.2c0 .21.15.46.55.38A8 8 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
            </svg>
          </a>
          <a href="#demo" className="nav-cta">
            Get early access
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
              <path d="M2 5.5h7m0 0L5.5 2M9 5.5L5.5 9" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </a>
        </div>
      </div>
    </nav>
  )
}
