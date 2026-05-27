import Nav from '@/components/Nav'
import Hero from '@/components/Hero'
import LogosStrip from '@/components/LogosStrip'
import Problem from '@/components/Problem'
import SpecShowcase from '@/components/SpecShowcase'
import ValueProps from '@/components/ValueProps'
import Quote from '@/components/Quote'
import Ecosystem from '@/components/Ecosystem'
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
      <SpecShowcase />
      <ValueProps />
      <Quote />
      <Ecosystem />
      <Blog />
      <FinalCTA />
      <Footer />
    </>
  )
}
