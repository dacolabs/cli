const painPoints = [
  {
    label: 'Deploy to verify',
    body: 'You push a schema change just to find out it breaks downstream. Every check needs a live warehouse.',
  },
  {
    label: 'No local test loop',
    body: 'There\'s no way to run pipeline logic on your laptop. You\'re testing in staging, if at all.',
  },
  {
    label: 'Fragmented definitions',
    body: 'The same table is described in dbt, a Notion doc, and a Slack thread. None of them match.',
  },
]

export default function Problem() {
  return (
    <section
      id="problem"
      style={{
        background: 'var(--paper-2)',
        padding: '96px 0',
        borderTop: '1px solid var(--line)',
        borderBottom: '1px solid var(--line)',
      }}
    >
      <div className="container">
        <div className="problem-cols" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '64px', alignItems: 'start' }}>
          <div>
            <div className="eyebrow">
              <span className="dot" />The problem
            </div>
            <h2 className="section-title" style={{ maxWidth: '16ch' }}>
              Your pipelines should run on your laptop.
            </h2>
            <p className="section-lede" style={{ marginBottom: 0 }}>
              Daco brings your data products into a single OpenDPI spec you can develop, test, and version locally.
              No deploy to verify a schema change. No warehouse connection required to run a test.
            </p>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px', paddingTop: '8px' }}>
            {painPoints.map((p) => (
              <div
                key={p.label}
                style={{
                  background: 'var(--paper)',
                  border: '1px solid var(--line)',
                  borderRadius: 'var(--radius)',
                  padding: '20px 24px',
                }}
              >
                <div style={{
                  fontWeight: 600,
                  fontSize: '0.9rem',
                  marginBottom: '6px',
                  color: 'var(--ink)',
                }}>
                  {p.label}
                </div>
                <div style={{ fontSize: '0.9rem', color: 'var(--muted)', lineHeight: 1.5 }}>
                  {p.body}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
