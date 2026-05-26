import Nav from '@/components/Nav'
import Hero from '@/components/Hero'
import LogosStrip from '@/components/LogosStrip'
import Problem from '@/components/Problem'
import ValueProps from '@/components/ValueProps'
import Ecosystem from '@/components/Ecosystem'
import SpecShowcase from '@/components/SpecShowcase'
import Quote from '@/components/Quote'
import Blog from '@/components/Blog'
import FinalCTA from '@/components/FinalCTA'
import Footer from '@/components/Footer'

export default function Home() {
  return (
    <>
      <Nav />
      <Hero />
      <LogosStrip />
      <Problem />
      <ValueProps />
      <Ecosystem />
      <SpecShowcase />
      <Quote />
      <Blog />
      <FinalCTA />
      <Footer />
    </>
  )
}
