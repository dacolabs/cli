export default function Problem() {
  return (
    <section
      id="problem"
      style={{
        background: 'var(--ink)',
        color: '#fff',
        padding: '96px 0',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* subtle grid — same as hero */}
      <div style={{
        position: 'absolute',
        inset: 0,
        backgroundImage: 'linear-gradient(rgba(255,255,255,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.04) 1px, transparent 1px)',
        backgroundSize: '64px 64px',
        maskImage: 'radial-gradient(ellipse 80% 60% at 50% 50%, #000 50%, transparent 100%)',
        pointerEvents: 'none',
      }} />

      <div className="container" style={{ position: 'relative', zIndex: 1 }}>
        <div className="eyebrow" style={{ color: 'rgba(255,255,255,0.45)' }}>
          <span className="dot" />The problem
        </div>
        <h2
          className="section-title"
          style={{ color: '#fff', maxWidth: '16ch' }}
        >
          Your pipelines should run on your laptop.
        </h2>
        <p
          className="section-lede"
          style={{ color: 'rgba(255,255,255,0.55)', marginBottom: 0 }}
        >
          Daco brings your data products into a single OpenDPI spec you can develop, test, and version locally.
          No deploy to verify a schema change. No warehouse connection required to run a test.
        </p>
      </div>
    </section>
  )
}
