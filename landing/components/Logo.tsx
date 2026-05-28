export default function Logo({ size = 26 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 100 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Daco"
    >
      <circle cx="50" cy="50" r="50" fill="#f4cf4a" />
      {/* stem — vertical bar of the d */}
      <rect x="65" y="14" width="14" height="72" rx="7" fill="#0c0c0c" />
      {/* bowl — circular counter of the d */}
      <circle cx="37" cy="63" r="22" fill="#0c0c0c" />
    </svg>
  )
}
