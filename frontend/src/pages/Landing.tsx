import { useAuth } from '@/hooks/useAuth'
import LandingNav from '@/components/landing/LandingNav'
import LandingHero from '@/components/landing/LandingHero'
import ClusterShowcase from '@/components/landing/ClusterShowcase'
import HowItWorks from '@/components/landing/HowItWorks'
import FeatureGrid from '@/components/landing/FeatureGrid'
import CtaStrip from '@/components/landing/CtaStrip'
import LandingFooter from '@/components/landing/LandingFooter'

export default function Landing() {
  const { user, isAuthenticated } = useAuth()

  return (
    <div id="top" className="min-h-svh bg-stone-50 text-stone-900">
      <LandingNav isAuthenticated={isAuthenticated} login={user?.login} />
      <LandingHero isAuthenticated={isAuthenticated} />
      <ClusterShowcase />
      <HowItWorks />
      <FeatureGrid />
      <CtaStrip isAuthenticated={isAuthenticated} />
      <LandingFooter />
    </div>
  )
}
