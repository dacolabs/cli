export default function LogosStrip() {
  const logos = [
    { name: 'Snowflake', shape: 'square' },
    { name: 'Databricks', shape: 'circle' },
    { name: 'BigQuery', shape: 'tri' },
    { name: 'dbt', shape: 'dia' },
    { name: 'Postgres', shape: 'circle' },
    { name: 'Kafka', shape: 'square' },
  ]

  return (
    <div className="logos">
      <div className="container logos-inner">
        <div className="logos-label">Works with the stack you already run</div>
        <div className="logos-row">
          {logos.map((l) => (
            <span key={l.name} className="logo-mock">
              <span className={`logo-mock-shape ${l.shape}`} />
              {l.name}
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}
