import Nav from '@/components/Nav'
import Hero from '@/components/Hero'
import Problem from '@/components/Problem'
import ValueProps from '@/components/ValueProps'
import Ecosystem from '@/components/Ecosystem'
import SpecShowcase from '@/components/SpecShowcase'
import Blog from '@/components/Blog'
import FinalCTA from '@/components/FinalCTA'
import Footer from '@/components/Footer'

export default function Home() {
  return (
    <>
      <Nav />
      <Hero />
      <Problem />
      <ValueProps />
      <Ecosystem />
      <SpecShowcase />
      <Blog />
      <FinalCTA />
      <Footer />
    </>
  )
}
