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
          <a href="#ecosystem">Product</a>
          <a href="#opendpi">OpenDPI</a>
          <a href="#docs">Docs</a>
          <a href="#blog">Blog</a>
          <a href="#community">Community</a>
        </div>
        <div className="nav-right">
          <a href="https://github.com/dacolabs/daco" target="_blank" rel="noopener noreferrer" style={{ fontSize: '14px', color: 'rgba(255,255,255,0.7)', fontWeight: 500 }}>
            GitHub ↗
          </a>
          <a href="#demo" className="nav-cta">
            Book a demo
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
              <path d="M2 5.5h7m0 0L5.5 2M9 5.5L5.5 9" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </a>
        </div>
      </div>
    </nav>
  )
}
