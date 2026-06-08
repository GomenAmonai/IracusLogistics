import { useState } from 'react'

import { Calculator, type CalcPrefill } from './components/Calculator'
import { Faq } from './components/Faq'
import { Footer } from './components/Footer'
import { Header } from './components/Header'
import { Hero } from './components/Hero'
import { HowItWorks } from './components/HowItWorks'
import { Kpi } from './components/Kpi'
import { LeadForm, type LeadPrefill } from './components/LeadForm'
import { Routes } from './components/Routes'
import { Services } from './components/Services'

function App() {
  const [prefill, setPrefill] = useState<LeadPrefill | null>(null)

  function handleSendAsLead(input: CalcPrefill) {
    setPrefill({
      toCity: input.toCity,
      cargoType: input.cargoType,
      weight: input.weight,
      volume: input.volume,
    })
    document.getElementById('lead')?.scrollIntoView({ behavior: 'smooth' })
  }

  return (
    <>
      <a
        href="#calc"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:inline-flex focus:items-center focus:rounded-full focus:bg-accent focus:px-5 focus:py-2.5 focus:text-sm focus:font-semibold focus:text-surface focus:shadow-card"
      >
        К калькулятору
      </a>
      <Header />
      <main>
        <Hero />
        <Kpi />
        <HowItWorks />
        <Calculator onSendAsLead={handleSendAsLead} />
        <Services />
        <Routes />
        <Faq />
        <LeadForm prefill={prefill} />
      </main>
      <Footer />
    </>
  )
}

export default App
