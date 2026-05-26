const props = [
  {
    num: '01',
    title: 'True local development',
    body: 'Build pipelines directly from your codebase, decoupled from data warehouses. Generate mock data, run tests, and translate to PySpark, dbt, or SQL on your laptop. No live connection required.',
  },
  {
    num: '02',
    title: 'Agent-direct AI integration',
    body: 'Claude Code and other agents use Daco to navigate live databases and metadata with your existing credentials. They pull directly from the systems you already trust, so pipelines build faster.',
  },
  {
    num: '03',
    title: 'Metadata-driven security',
    body: 'Engineers and agents pull schemas and metadata. They never pull rows. Sensitive data stays in your warehouse, and you skip the access management layer a normal catalog would need.',
  },
]

export default function ValueProps() {
  return (
    <section>
      <div className="container">
        <div className="eyebrow"><span className="dot" />Core advantages</div>
        <h2 className="section-title">What changes with Daco.</h2>
        <p className="section-lede">Define your data products once. Develop locally. Let AI agents work with metadata instead of data.</p>
        <div className="value-grid">
          {props.map((p) => (
            <div key={p.num} className="value-card">
              <span className="value-num">{p.num}</span>
              <h3>{p.title}</h3>
              <p>{p.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
