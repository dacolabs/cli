const props = [
  {
    num: '01',
    title: 'Edit, run, repeat. Offline.',
    body: 'Validate schemas, generate fixtures, run pipeline tests against mock data. No warehouse connection required to know your change is correct.',
  },
  {
    num: '02',
    title: 'One spec, every target.',
    body: 'Write your data product once in OpenDPI. Translate to PySpark, dbt, SQL, Pydantic, Go, or TypeScript types, regenerated on every change.',
  },
  {
    num: '03',
    title: 'AI agents in the loop.',
    body: 'Claude Code reads your live schemas through Daco, edits the spec, and regenerates types. Agents work from metadata, never the data itself.',
  },
]

export default function ValueProps() {
  return (
    <section>
      <div className="container">
        <div className="eyebrow"><span className="dot" />Local development</div>
        <h2 className="section-title">Local-first, by design.</h2>
        <p className="section-lede">One spec. Your editor. Your stack. No live warehouse to test a schema change.</p>
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
