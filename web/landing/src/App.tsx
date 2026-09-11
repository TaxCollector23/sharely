import { Header } from './components/Header'
import { Hero } from './components/Hero'
import { HowItWorks } from './components/HowItWorks'
import { Examples } from './components/Examples'
import { DevServer } from './components/DevServer'
import { Security } from './components/Security'
import { Cli } from './components/Cli'
import { Install } from './components/Install'
import { Footer } from './components/Footer'

function App() {
  return (
    <div className="min-h-screen bg-[var(--bg)] text-[var(--text)]">
      <Header />
      <main>
        <Hero />
        <HowItWorks />
        <Examples />
        <DevServer />
        <Security />
        <Cli />
        <Install />
      </main>
      <Footer />
    </div>
  )
}

export default App
